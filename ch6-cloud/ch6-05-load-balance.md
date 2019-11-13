# 6.5 Load Balancing

This section will discuss common distributed system load balancing methods.

## 6.5.1 Common Load Balancing Ideas

If we don't consider the equilibrium, there are now n service nodes, and we only need to pick one of the n from the business process. There are several ideas:

1. Pick in order: For example, the last time you selected the first one, then this time you choose the second one, the next one, if you have already reached the last one, then the next one starts from the first one. In this case, we can store the service node information in an array. After each request is completed downstream, we can move an index back. Move back to the beginning of the array when you move to the end.

2. Pick one randomly: Pick each time randomly, and be random and pseudo-random. Assuming that the xth machine is selected, then x can be described as `rand.Intn()%n`.

3. Sort the downstream nodes according to a certain weight and select the one with the largest/small weight.

Of course, the actual scenario we can't have no brain polling or no brain randomness. If the downstream request fails, we still need some mechanism to retry. If there is a pure random algorithm, there is a certain possibility that you will be next time. Still random to this problem node.

Let's look at a load balancing case for a production environment.

## 6.5.2 Load Balancing Based on Shuffle Algorithm

Considering that we need to randomly select the node that sends the request each time, and try to retry the other nodes when it encounters a downstream return error. So we design an index array with the same size and node array size. Every time we come to a new request, we shuffle the index array, then take the first element as the selected service node. If the request fails, then select the next node. Try again, and so on:

```go
Var endpoints = []string {
"100.69.62.1:3232",
"100.69.62.32: 3232",
"100.69.62.42: 3232",
"100.69.62.81:3232",
"100.69.62.11:3232",
"100.69.62.113:3232",
"100.69.62.101:3232",
}

// focus on this shuffle
Func shuffle(slice []int) {
For i := 0; i < len(slice); i++ {
a := rand.Intn(len(slice))
b := rand.Intn(len(slice))
Slice[a], slice[b] = slice[b], slice[a]
}
}

Func request(params map[string]interface{}) error {
Var indexes = []int {0,1,2,3,4,5,6}
Var err error

Shuffle(indexes)
maxRetryTimes := 3

Idx := 0
For i := 0; i < maxRetryTimes; i++ {
Err = apiRequest(params, indexes[idx])
If err == nil {
Break
}
Idx++
}

If err != nil {
// logging
Return err
}

Return nil
}
```

We loop through the slices, swapping them in pairs, which is similar to the shuffling method we usually use when playing cards. It looks like there is no problem.

### 6.5.2.1 Unbalanced load caused by incorrect shuffling

Really no problem? Still have problems. There are two hidden pitfalls in this short program:

1. There are no random seeds. In the absence of a random seed, the sequence of pseudo-random numbers returned by `rand.Intn()` is fixed.

2. Uneven shuffle, which will cause the first node of the entire array to have a high probability of being selected, and the load distribution of multiple nodes is not balanced.

The first point is relatively simple and should not be given proof here. Regarding the second point, we can use the knowledge of probability to simply prove it. Assuming that each selection is truly random, we assume that the probability that the node at the first position is not selected in the `len(slice)` exchange is `((6/7)*(6/7))^7 ≈ 0.34`. In the case of uniform distribution, we definitely want the probability that the first element is distributed at any position is equal, so the probability of being randomly selected should be approximately equal to `1/7≈0.14`.

Obviously, the shuffling algorithm given here has a 30% probability of not swapping elements for any position. So all elements tend to stay in their original position. Because every time we input the same sequence for the `shuffle` array, the first element has a higher probability of being selected. In the case of load balancing, it means that the first machine load in the node array will be much higher than other machines (at least 3 times more).

### 6.5.2.2 Correcting the shuffling algorithm

The mathematically proven fisher-yates algorithm is the main idea of ​​picking a value at random and placing it at the end of the array. Then randomly pick a value in the array of n-1 elements, put it at the end of the array, and so on.

```go
Func shuffle(indexes []int) {
For i:=len(indexes); i>0; i-- {
lastIdx := i - 1
Idx := rand.Int(i)
Indexes[lastIdx], indexes[idx] = indexes[idx], indexes[lastIdx]
}
}
```

The algorithm has been built for us in Go's standard library:

```go
Func shuffle(n int) []int {
b := rand.Perm(n)
Return b
}
```

In the current scenario, we can use `rand.Perm` to get the index array we want.

## 6.5.3 Random node selection problem for ZooKeeper cluster

The scenario in this section is to select a node from N nodes to send a request. After the initial request is over, subsequent requests will reshuffle the array, so there is no relationship between every two requests. Therefore, our shuffling algorithm above does not theoretically initialize the seeds of the random library.

However, in some special scenarios, such as when using ZooKeeper, when the client initiates a node selection from multiple service nodes, a long connection is established to the node. The client request is then sent to the node. The next node is selected in the node list until the node is unavailable. In this scenario, our initial connection node selection requires that it be "true" random. Otherwise, all clients will connect to the same instance of ZooKeeper when they start, and they will not be able to load balance. If your business is similar in your daily development, it's important to consider whether a similar situation will occur. How to set the seed for the rand library:

```go
rand.Seed(time.Now().UnixNano())
```

The reason for these conclusions is that the earlier version of a widely used open source ZooKeeper library made the above mistakes, and it was not until early 2016 that the problem was fixed.

## 6.5.4 Load balancing algorithm effect verification

We do not consider the case of weighted load balancing here, since the name is the load "equalization". Then the most important thing is balance. We simply compare the shuffle algorithm in the opening with the results of the fisher yates algorithm:

```go
Package main

Import (
"fmt"
"math/rand"
"time"
)

Func init() {
rand.Seed(time.Now().UnixNano())
}

Func shuffle1(slice []int) {
For i := 0; i < len(slice); i++ {
a := rand.Intn(len(slice))
b := rand.Intn(len(slice))
Slice[a], slice[b] = slice[b], slice[a]
}
}

Func shuffle2(indexes []int) {
For i := len(indexes); i > 0; i-- {
lastIdx := i - 1
Idx := rand.Intn(i)
Indexes[lastIdx], indexes[idx] = indexes[idx], indexes[lastIdx]
}
}

Func main() {
Var cnt1 = map[int]int{}
For i := 0; i < 1000000; i++ {
Var sl = []int{0, 1, 2, 3, 4, 5, 6}
Shuffle1(sl)
Cnt1[sl[0]]++
}

Var cnt2 = map[int]int{}
For i := 0; i < 1000000; i++ {
Var sl = []int{0, 1, 2, 3, 4, 5, 6}
Shuffle2(sl)
Cnt2[sl[0]]++
}

fmt.Println(cnt1, "\n", cnt2)
}
```

Output:

```shell
Map[0:224436 1:128780 5:129310 6:129194 2:129643 3:129384 4:129253]
Map[6:143275 5:143054 3:143584 2:143031 1:141898 0:142631 4:142527]
```

The distribution results are consistent with the conclusions we have derived.