# 6.5 Load-Balance 负载均衡

本节将会讨论常见的分布式系统负载均衡手段。

## 6.5.1 常见的负载均衡思路

如果我们不考虑均衡的话，现在有 n 个 endpoint，我们完成业务流程实际上只需要从这 n 个中挑出其中的一个。有几种思路:

1. 按顺序挑: 例如上次选了第一台，那么这次就选第二台，下次第三台，如果已经到了最后一台，那么下一次从第一台开始。这种情况下我们可以把 endpoint 都存储在数组中，每次请求完成下游之后，将一个索引后移即可。在移到尽头时再移回数组开头处。

2. 随机挑一个: 每次都随机挑，真随机伪随机均可。假设选择第 x 台机器，那么 x 可描述为 `rand.Intn() % n`。

3. 根据某种权重，对下游 endpoints 进行排序，选择权重最大/小的那一个。

当然了，实际场景我们不可能无脑轮询或者无脑随机，如果对下游请求失败了，我们还需要某种机制来进行重试，如果纯粹的随机算法，存在一定的可能性使你在下一次仍然随机到这次的问题节点。

我们来看一个生产环境的负载均衡案例。

## 6.5.2 基于洗牌算法的负载均衡

考虑到我们需要随机选取每次发送请求的 endpoint，同时在遇到下游返回错误时换其它节点重试。所以我们设计一个大小和 endpoints 数组大小一致的索引数组，每次来新的请求，我们对索引数组做洗牌，然后取第一个元素作为选中的服务节点，如果请求失败，那么选择下一个节点重试，以此类推:

```go
var endpoints = []string {
    "100.69.62.1:3232",
    "100.69.62.32:3232",
    "100.69.62.42:3232",
    "100.69.62.81:3232",
    "100.69.62.11:3232",
    "100.69.62.113:3232",
    "100.69.62.101:3232",
}

// 重点在这个 shuffle
func shuffle(slice []int) {
    for i := 0; i < len(slice); i++ {
        a := rand.Intn(len(slice))
        b := rand.Intn(len(slice))
        slice[a], slice[b] = slice[b], slice[a]
    }
}

func request(params map[string]interface{}) error {
    var indexes = []int {0,1,2,3,4,5,6}
    var err error

    shuffle(indexes)
    maxRetryTimes := 3

    idx := 0
    for i := 0; i < maxRetryTimes; i++ {
        err = apiRequest(params, indexes[idx])
        if err == nil {
            break
        }
        idx++
    }

    if err != nil {
        // logging
        return err
    }

    return nil
}

```

我们循环一遍 slice，两两交换，这个和我们平常打牌时常用的洗牌方法类似。看起来没有什么问题。

### 6.5.2.1 错误的洗牌导致的负载不均衡

真的没有问题么？实际上还是有问题的。这段简短的程序里有两个隐藏的隐患:

1. 没有随机种子。在没有随机种子的情况下，rand.Intn 返回的伪随机数序列是固定的。

2. 洗牌不均匀，会导致整个数组第一个节点有大概率被选中，并且多个节点的负载分布不均衡。

第一点比较简单，应该不用在这里给出证明了。关于第二点，我们可以用概率知识来简单证明一下。假设每次挑选都是真随机，我们假设第一个位置的 endpoint 在 len(slice) 次交换中都不被选中的概率是 ((6/7)*(6/7))^7 ≈ 0.34。而分布均匀的情况下，我们肯定希望被第一个元素在任意位置上分布的概率均等，所以其被随机选到的概率应该 ≈ 1/7 ≈ 0.14。

显然，这里给出的洗牌算法对于任意位置的元素来说，有 30% 的概率不对其进行交换操作。所以所有元素都倾向于留在原来的位置。因为我们每次对 shuffle 数组输入的都是同一个序列，所以第一个元素有更大的概率会被选中。在负载均衡的场景下，也就意味着 endpoints 数组中的第一台机器负载会比其它机器高不少(这里至少是 3 倍以上)。

### 6.5.2.2 修正洗牌算法

从数学上得到过证明的还是经典的 fisher-yates 算法，主要思路为每次随机挑选一个值，放在数组末尾。然后在 n-1 个元素的数组中再随机挑选一个值，放在数组末尾，以此类推。

```go
func shuffle(indexes []int) {
    for i:=len(indexes); i>0; i-- {
        lastIdx := i - 1
        idx := rand.Int(i)
        indexes[lastIdx], indexes[idx] = indexes[idx], indexes[lastIdx]
    }
}
```

在 Go 的标准库中实际上已经为我们内置了该算法:

```go
func shuffle(n int) []int {
    b := rand.Perm(n)
    return b
}
```

在当前的场景下，我们只要用 rand.Perm 就可以得到我们想要的索引数组了。

## 6.5.3 zk 集群的随机节点挑选问题

本节中的场景是从 N 个节点中选择一个节点发送请求，初始请求结束之后，后续的请求会重新对数组洗牌，所以每两个请求之间没有什么关联关系。因此我们上面的洗牌算法，理论上不初始化随机库的种子也是不会出什么问题的。

但在一些特殊的场景下，例如使用 zk 时，客户端初始化从多个服务节点中挑选一个节点后，是会向该节点建立长连接的。并且之后如果有请求，也都会发送到该节点去。直到该节点不可用，才会在 endpoints 列表中挑选下一个节点。在这种场景下，我们的初始连接节点选择就要求必须是“真”随机了。否则，所有客户端起动时，都会去连接同一个 zk 的实例，根本无法起到负载均衡的目的。如果在日常开发中，你的业务也是类似的场景，也务必考虑一下是否会发生类似的情况。为 rand 库设置种子的方法:

```go
rand.Seed(time.Now().UnixNano())
```

之所以会有上面这些结论，是因为某个使用较广泛的开源 zk 库的早期版本就犯了上述错误，直到 2016 年早些时候，这个问题才被修正。

## 6.5.4 负载均衡算法效果验证

我们这里不考虑加权负载均衡的情况，既然名字是负载“均衡”。那么最重要的就是均衡。我们把开篇中的 shuffle 算法，和之后的 fisher yates 算法的结果进行简单地对比：

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

func shuffle1(slice []int) {
    for i := 0; i < len(slice); i++ {
        a := rand.Intn(len(slice))
        b := rand.Intn(len(slice))
        slice[a], slice[b] = slice[b], slice[a]
    }
}

func shuffle2(indexes []int) {
    for i := len(indexes); i > 0; i-- {
        lastIdx := i - 1
        idx := rand.Intn(i)
        indexes[lastIdx], indexes[idx] = indexes[idx], indexes[lastIdx]
    }
}

func main() {
    var cnt1 = map[int]int{}
    for i := 0; i < 1000000; i++ {
        var sl = []int{0, 1, 2, 3, 4, 5, 6}
        shuffle1(sl)
        cnt1[sl[0]]++
    }

    var cnt2 = map[int]int{}
    for i := 0; i < 1000000; i++ {
        var sl = []int{0, 1, 2, 3, 4, 5, 6}
        shuffle2(sl)
        cnt2[sl[0]]++
    }

    fmt.Println(cnt1, "\n", cnt2)
}

```

输出：

```shell
map[0:224436 1:128780 5:129310 6:129194 2:129643 3:129384 4:129253]
map[6:143275 5:143054 3:143584 2:143031 1:141898 0:142631 4:142527]
```

分布结果和我们推导出的结论是一致的。
