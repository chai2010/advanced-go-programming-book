# 6.3 Delay Mission System

When we are doing the system, we often deal with real-time tasks. The request is processed immediately, and then the user is immediately given feedback. But sometimes you will encounter non-real-time tasks, such as issuing important announcements at certain points in time. Or you need to do something specific after the user has done something X minutes / Y hours, such as notifications, issuing bonds, and so on.

If the business scale is relatively small, sometimes we can also use the database to coordinate with the polling to simply handle this task, but companies of scale will naturally find more versatile solutions to solve this kind of problem.

There are generally two ways to solve this problem:

1. Implement a distributed timing task management system similar to crontab.
2. Implement a message queue that supports scheduled messages.

The two ideas have led to a number of different systems, but the essence is similar. It is necessary to implement a timer. In the stand-alone scenario, the timer is not uncommon. For example, when we deal with the network library, we often call the `SetReadDeadline()` function, which creates a timer locally. After the specified time, we will receive it. A notification to the timer tells us that the time has arrived. At this time, if the reading has not been completed, it can be considered that a network problem has occurred, thereby interrupting the reading.

Let's start with the timer and explore the implementation of the delayed task system.

## 6.3.1 Implementation of the timer

The implementation of timers has been a problem in the industry. Common is the time heap and time wheel.

### 6.3.1.1 Time heap

The most common time heap is usually implemented with a small top heap. The small top heap is actually a special binary tree. See *Figure 6-4*

![二叉堆](../images/ch6-binary_tree.png)

*Figure 6-4 Binary Heap Structure*

What are the benefits of a small top heap? For timers, if the top element is larger than the current time, then all elements in the heap are larger than the current time. Furthermore, we have no need to deal with the time heap at this moment. The time complexity of the timing check is `O(1)`.

When we find that the elements at the top of the heap are smaller than the current time, then it may be that a batch of events has already begun to expire, and normal pop-up and heap adjustment operations are fine. The time complexity of each heap adjustment is `O(LgN)`.

Go's own built-in timer is implemented with a time heap, but instead of using a binary heap, a flatter quad fork is used. In the most recent version, some optimizations have been added. Let's not talk about optimization first. Let's first look at what the four-prong small top stack looks like:

![Quad fork] (../images/ch6-four-branch-tree.png)

*Figure 6-5 Quad Cross Stack Structure*

The nature of the small top heap, the parent node is smaller than its four child nodes, there is no special size relationship between the child nodes.

There is no essential difference between element timeout and heap adjustment in a quad fork heap and a binary heap.

### 6.3.1.2 Time Wheel

![timewheel](../images/ch6-timewheel.png)

*Figure 6-6 Time Wheel*

When using the time wheel to implement the timer, we need to define the "scale" of each grid. The time wheel can be imagined as a clock, and the center has a second hand clockwise. Each time we turn to a scale, we need to see if the task list mounted on that scale has a task that has expired.

Structurally, the time wheel is similar to the hash table, if we define the hash algorithm as: trigger time % time wheel element size. Then this is a simple hash table. In the case of a hash collision, a timer for hooking the hash collision is used.

In addition to this single-layer time wheel, there are some time wheels in the industry that use multiple layers of implementation, so I won't go into details here.

## 6.3.2 Task Distribution

With a basic timer implementation, if we are developing a stand-alone system, we can pick up the sleeves, but in this chapter we are talking about distributed, and there is still some distance from the "distributed".

We also need to distribute these "timing" or "delay" (essentially timing) tasks. Here is an idea:

![task-dist](../images/ch6-task-sched.png)

*Figure 6-7 Distributed Task Distribution*

Every hour, every instance, will go to the database to retrieve the scheduled tasks that need to be processed in the next hour. Just pick those tasks with `task_id % shard_count = shard_id`.

When these timing tasks are triggered, you need to notify the user side. There are two ways to do this:

1. Encapsulate the information triggered by the task as a message and send it to the message queue. The user side listens to the message queue.
2. Call the user-configured callback function.

Both schemes have their own advantages and disadvantages. If you use 1, then if the message queue fails, the whole system will be unavailable. Of course, the current message queue will generally have its own high-availability solution. Most of the time, we don't have to worry about this problem. . Secondly, if the message queue is used in the middle of the general business process, the delay will increase. If the timed task must be completed within tens of milliseconds to several hundred milliseconds after the trigger, then the message queue will have certain risks. If you adopt 2, it will increase the burden of the timing task system. We know that the most feared execution of a single-machine timer is that the callback function takes too long to execute, which will block subsequent task execution. In a distributed scenario, this concern is still applicable. An irresponsible business callback may directly drag down the entire timed mission system. Therefore, we also need to consider adding a tested timeout setting based on the callback, and carefully review the timeout period filled in by the user.

## 6.3.3 Data rebalancing and idempotency considerations

When our task performs a cluster machine failure, the task needs to be reallocated. According to the previous modulo strategy, it is more troublesome to redistribute tasks that have not been processed by this machine. If it is an online system that is actually running, you should pay more attention to the task balance in the event of a failure.

Here is an idea:

We can refer to Elasticsearch's data distribution design, each task data has multiple copies, here assume two copies, as shown in Figure 6-8*:

![Data Distribution](../images/ch6-data-dist1.png)

*Figure 6-8 Task Data Distribution*

Although there is two holders of a piece of data, the copy held by the holder will be distinguished, such as whether it is a master copy or a non-master copy. The master copy is a black part in the figure, and the non-master copy is a normal line. .

A task will only be executed on the node holding the master copy.

When the machine fails, the task data needs to work on data rebalancing. For example, node 1 is hung up, see *Figure 6-9*.

![Data Distribution 2](../images/ch6-data-dist2.png)

*Figure 6-9 Data distribution at fault*

The data of node 1 will be migrated to node 2 and node 3.

Of course, you can also use a slightly more complicated idea, such as the role division of nodes in the cluster, and the coordination node to do the task redistribution work in this failure. Considering the high availability, the coordination node may also need 1 to 2 An alternate node to prevent accidents.

As mentioned before, we will use the message queue to trigger the notification to the user. When using the message queue, many queues do not support the semantics of `exactly once`. In this case, we need to let the user take care of the deduplication or consumption of the message. Idempotent processing.