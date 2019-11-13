# 5.9 Grayscale Publishing and A/B test

Medium-sized Internet companies tend to have millions of users, while large Internet companies' systems may have to serve tens of millions or even billions of users. The inflow of requests from large systems is often endless, and any wind and grass will surely be felt by end users. For example, if your system refuses some upstream requests on the way to the line, and this time depends on your system without any fault tolerance, then this error will always be thrown up until it reaches the end user. Form a real damage to the user. This kind of damage may be a strange string that pops up on the user's app so that the user can't figure it out. Users can forget this by refreshing the page. But it may also make users who are eager to grab the spiked products with tens of thousands of competitors at the same time, because of the small problem in the code, lost the first-mover advantage, and missed the favorite products that they have spent a few months. How much damage is done to the user depends on how important your system is to your users.

In any case, fault tolerance is important in large systems, and it is important to have the system reach the end user in percentages and in batches. Although today's Internet company systems nominally say that they have been thoroughly and rigorously tested before going online, even if they do, code bugs are always inevitable. Even if the code is free of bugs, collaboration between distributed services can be non-technical in terms of "logic."

At this time, the grayscale release is very important. The grayscale release is also called the canary release. It is said that the 17th century British mine workers found that the canary is very sensitive to gas. When the gas reaches a certain concentration, the canary is Will die, but the canary's lethal gas is not fatal to people, so the canary is used as their gas detection tool. Grayscale publishing of Internet systems is generally achieved in two ways:

1. Implement grayscale publishing through batch deployment
2. Grayscale publishing through business rules

The first method uses more when iterating over the old functions of the system. When the new function is online, the second method is used more. Of course, when making major changes to the more important old functions, it is generally preferred to publish them according to business rules, because the risk of direct full opening to all users is too great.

## 5.9.1 Implementing grayscale publishing through batch deployment

If the service is deployed on 15 instances (which may be physical machines or containers), we divide the 15 instances into four groups. In order of priority, there are 1-2-4-8 machines, each time. When expanding, it is probably twice as large.

![online group](../images/ch5-online-group.png)

*Figure 5-20 Group Deployment*

Why use 2 times? This will ensure that we will not divide the group too much, no matter how many machines we have. For example, 1024 machines, only need to deploy 1-2-4-8-16-32-64-128-256-512 ten times, you can deploy all.

In this way, the users that we first affected on the line accounted for a small proportion of the total users, such as the service of 1000 machines. If there is a problem after we go online, it will only affect 1/1000 users. If 10 groups are completely evenly divided, that online will affect 1/10 users immediately, and 1/10 of the business problems, it may be an irreparable accident for the company.

When going online, the most effective way to observe is to look at the error log of the program. If there is a more obvious logic error, the scroll speed of the error log will increase with the naked eye. These errors can also be reported to the monitoring system in the company through a system such as metrics. Therefore, during the online process, it is also possible to observe whether the abnormality occurs by observing the monitoring curve.

If there is an abnormal situation, the first thing to do is to roll back.

## 5.9.2 Grayscale publishing through business rules

There are many common grayscale strategies, and simpler requirements. For example, our strategy is to publish in thousands of points. Then we can use user id, mobile phone number, user device information, etc. to generate a simple hash value. And then ask for the model, using pseudo code to indicate:

```go
// pass 3/1000
Func passed() bool {
Key := hashFunctions(userID) % 1000
If key <= 2 {
Return true
}

Return false
}
```

### 5.9.2.1 Optional Rules

Common grayscale publishing systems have the following rules to choose from:

1. Published by city
2. Publish by probability
3. Publish by percentage
4. Publish by whitelist
5. Released by line of business
6. Publish by UA (APP, Web, PC)
7. Publish by distribution channel

Because it is related to the company's business, cities, lines of business, UAs, distribution channels, etc., may be directly encoded in the system, but the functions are similar.

Publishing by whitelist is relatively simple. When the function is online, we hope that only employees and testers inside the company can access new functions. They will directly write accounts and mailboxes to the whitelist and refuse access to any other accounts.

Publishing by probability means implementing a simple function:

```go
Func isTrue() bool {
Return true/false according to the rate provided by user
}
```

It can return `true` or `false` according to the probability specified by the user. Of course, the probability of `true` plus the probability of `false` should be 100%. This function does not require any input.

Publishing by percentage means implementing a function like this:

```go
Func isTrue(phone string) bool {
If hash of phone matches {
Return true
}

Return false
}
```

This situation can return the corresponding `true` and `false` according to the specified percentage, and the above simple difference according to probability is that we need the caller to provide us with an input parameter, we use the input parameter as the source to calculate Hash, and try to find the result after hashing, and return the result. This ensures that the same user's return result is consistent with multiple calls. In the following scenarios, the grayscale algorithm that can be expected for this result must be used, as shown in Figure 5-21*.

![set and get processes should not go to different versions of the API because of grayscale] (../images/ch5-set-time-line.png)

*Figure 5-21 Set first and then get*

If a random strategy is used, problems such as *Figure 5-22* may occur:

![set and get processes should not go to different versions of the API because of grayscale] (../images/ch5-set-time-line_2.png)

*Figure 5-22 Set first and then get*

For a specific example, the registration part of the website may have two sets of APIs, which are grayscale according to the user ID, which are different access logics. If the V1 version of the API is used for storage and the V2 version of the API is used for the acquisition, then there may be a strange problem of returning the registration failure message after the user successfully registers.

## 5.9.3 How to implement a set of grayscale publishing system

As mentioned earlier, the interface provided to the user can be roughly divided into simple grayscale judgment logic bound to the service. And enter a slightly more complicated hash grayscale. Let's take a look at how to implement such a grayscale system (function).

### 5.9.3.1 Business-related simple grayscale

The company generally has a mapping relationship between public city names and ids. If the business only involves China, the number of cities will not be particularly large, and the ids may all be within 10,000. Then we just need to open a bool array of about 10,000 sizes to meet the needs:

```go
Var cityID2Open = [12000]bool{}

Func init() {
readConfig()
For i:=0;i<len(cityID2Open);i++ {
If city i is opened in configs {
cityID2Open[i] = true
}
}
}

Func isPassed(cityID int) bool {
Return cityID2Open[cityID]
}
```

If the company assigns a larger value to cityID, then we can consider using map to store the mapping. The map query is slightly slower than the array, but the extension will be more flexible:

```go
Var cityID2Open = map[int]struct{}{}

Func init() {
readConfig()
For _, city := range openCities {
cityID2Open[city] = struct{}{}
}
}

Func isPassed(cityID int) bool {
If _, ok := cityID2Open[cityID]; ok {
Return true
}

Return false
}
```

According to the white list, by business line, by UA, by distribution channel, it is essentially the same as the release by city, and will not be described here.

Publishing by probability is a bit more special, but it's easy to implement without regard to input:

```go

Func init() {
rand.Seed(time.Now().UnixNano())
}

// rate is 0~100
Func isPassed(rate int) bool {
If rate >= 100 {
Return true
}

If rate > 0 && rand.Int(100) > rate {
Return true
}

Return false
}
```

Note the initialization seed.

### 5.9.3.2 Hash Algorithm

There are many algorithms for hashing, such as md5, crc32, sha1, etc., but our purpose here is just to map these data, and I don’t want to use too many CPUs because of the hash calculation. The more algorithms are murmurhash, here is our simple benchmark for these common hash algorithms.

The following uses the standard library md5, sha1 and open source murmur3 implementation for comparison.

```go
Package main

Import (
"crypto/md5"
"crypto/sha1"

"github.com/spaolacci/murmur3"
)

Var str = "hello world"

Func md5Hash() [16]byte {
Return md5.Sum([]byte(str))
}

Func sha1Hash() [20]byte {
Return sha1.Sum([]byte(str))
}

Func murmur32() uint32 {
Return murmur3.Sum32([]byte(str))
}

Func murmur64() uint64 {
Return murmur3.Sum64([]byte(str))
}
```

Write a benchmark test for these algorithms:

```go
Package main

Import "testing"

Func BenchmarkMD5(b *testing.B) {
For i := 0; i < b.N; i++ {
md5Hash()
}
}

Func BenchmarkSHA1(b *testing.B) {
For i := 0; i < b.N; i++ {
sha1Hash()
}
}

Func BenchmarkMurmurHash32(b *testing.B) {
For i := 0; i < b.N; i++ {
Murmur32()
}
}

Func BenchmarkMurmurHash64(b *testing.B) {
For i := 0; i < b.N; i++ {
Murmur64()
}
}

```

Then look at the running effect:

```shell
~/t/g/hash_bench git:master ❯❯❯ go test -bench=.
Goos: darwin
Goarch: amd64
BenchmarkMD5-4 10000000 180 ns/op
BenchmarkSHA1-4 10000000 211 ns/op
BenchmarkMurmurHash32-4 50000000 25.7 ns/op
BenchmarkMurmurHash64-4 20000000 66.2 ns/op
PASS
Ok _/Users/caochunhui/test/go/hash_bench 7.050s
```

Visible muRmur hash has more than three times the performance improvement over other algorithms. Obviously, to do load balancing, murmurhash is better than md5 and sha1. In the past few years, there are other more efficient hashing algorithms in the community. Interested readers can investigate.

### 5.9.3.3 Is the distribution uniform?

For the hash algorithm, in addition to performance issues, it is also necessary to consider whether the hashed values ​​are evenly distributed. If the value of the hash is not evenly distributed, it will naturally not achieve the effect of uniform gray.

Take murmur3 as an example. Let's start with 15810000000, make 10 million numbers similar to the mobile phone number, then divide the calculated hash value into ten buckets and observe whether the count is even:

```go
Package main

Import (
"fmt"

"github.com/spaolacci/murmur3"
)

Var bucketSize = 10

Func main() {
Var bucketMap = map[uint64]int{}
For i := 15000000000; i < 15000000000+10000000; i++ {
hashInt := murmur64(fmt.Sprint(i)) % uint64(bucketSize)
bucketMap[hashInt]++
}
fmt.Println(bucketMap)
}

Func murmur64(p string) uint64 {
Return murmur3.Sum64([]byte(p))
}
```

Take a look at the execution results:

```shell
Map[7:999475 5:1000359 1:999945 6:1000200 3:1000193 9:1000765 2:1000044 \
4:1000343 8:1000823 0:997853]
```

The deviation is within 1/100 and is acceptable. When the reader is investigating other algorithms and judging whether it can be used for grayscale publishing, it should also examine the performance and balance mentioned in this section.