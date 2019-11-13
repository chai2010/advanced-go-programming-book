# 6.7 Distributed crawler

The information explosion in the Internet era is a problem that many people feel headaches. The news, information, and videos that are inexhaustible are invading our fragmentation time. But on the other hand, when we really need data, we feel that the data is not so easy to obtain. For example, we want to analyze what people are discussing and what they care about now. Sometimes, maybe we just don't have time to read the favorite novels one by one, but we want to use technology to put them in our own database. Look back even if it is a few months or a year later. Or maybe we want to save these fleeting and useful information on the Internet, such as the high-quality discussions of the good people gathered in a very small forum, at some point in the future, even if these small gathering areas are not Thought that it will continue, let us pull out the original precious point of view from the hard disk.

In addition to the emotional needs, there are a lot of precious open materials on the Internet. In recent years, deep learning has been hot and hot, but machine learning is often not suffering from whether my model is properly established, whether my parameters are adjusted correctly, but rather Initial start-up phase: no data.

As a pre-work to collect data, the ability to write a simple or complex crawler is still very important to us.

## 6.7.1 Stand-alone crawler based on collly

"Go Language Programming" gives a simple crawler example. After years of development, it is more convenient to use Go to write a crawler for a website, such as using crawl to crawl a website (virtual site, here abcdefg As a placeholder) in the first ten pages of the Go language tag:

```go
Package main

Import (
"fmt"
"regexp"
"time"

"github.com/gocolly/colly"
)

Var visited = map[string]bool{}

Func main() {
// Instantiate default collector
c := colly.NewCollector(
colly.AllowedDomains("www.abcdefg.com"),
colly.MaxDepth(1),
)

// We think the matching page is the details page of the site
detailRegex, _ := regexp.Compile(`/go/go\?p=\d+$`)
// Matching the following pattern is the list page of the site
listRegex, _ := regexp.Compile(`/t/\d+#\w+`)

// All a tags, set the callback function
c.OnHTML("a[href]", func(e *colly.HTMLElement) {
Link := e.Attr("href")

// Visited details page or list page, skipped
If visited[link] && (detailRegex.Match([]byte(link)) || listRegex.Match([]byte(link))) {
Return
}

// is neither a list page nor a detail page
// Then it's not what we care about, skip it
If !detailRegex.Match([]byte(link)) && !listRegex.Match([]byte(link)) {
Println("not match", link)
Return
}

// Because most websites have anti-reptile strategies
// So there should be sleep logic in the crawler logic to avoid being blocked
time.Sleep(time.Second)
Println("match", link)

Visited[link] = true

time.Sleep(time.Millisecond * 2)
c.Visit(e.Request.AbsoluteURL(link))
})

Err := c.Visit("https://www.abcdefg.com/go/go")
If err != nil {fmt.Println(err)}
}
```

## 6.7.2 Distributed crawler

Imagine that your information analysis system is running very fast. The speed of getting information has become a bottleneck. Although you can use all the excellent concurrency features of the Go language to fully use the CPU and network bandwidth of a single machine, you still want to speed up the crawling speed of the crawler. In many scenarios, speed makes sense:

1. For e-commerce during the price war, I hope to get the latest price after the opponent's price changes, and then automatically adjust the price of the product by the machine.
2. For feed stream services like headlines, the timeliness of information is also very important. If the news we crawled slowly was yesterday's news, it wouldn't make any sense to the user.

So we need distributed crawlers. In essence, distributed crawlers are a set of task distribution and execution systems. The common task distribution, because there is a speed mismatch between the upstream and downstream, must rely on the message queue.

![dist-crawler](../images/ch6-dist-crawler.png)

*Figure 6-14 Reptile Workflow*

The main job of the upstream is to crawl all the target "list pages" according to the pre-configured starting point. The html content of the list page will contain links to all the detail pages. The number of detail pages is typically 10 to 100 times that of the list page, so we link these details pages as "tasks" and distribute them through the message queue.

For page crawling, it's not really important to have occasional repetitions during execution, because the task results are idempotent (here we only crawl the page content, not the comments section).

In this section, we will simply implement a message queue-based crawler. In this section, we use nats for task distribution. In actual development, the reliability requirements of the message itself and the company's infrastructure components should be selected for their own business.

### 6.7.2.1 Introduction to nats

Nats is a high-performance distributed message queue implemented by Go for high-concurrency, high-throughput messaging scenarios. Early nats were speed-oriented and did not support persistence. Since 16 years, nats has supported log-based persistence through nats-streaming, as well as reliable messaging. For the sake of demonstration, we only use nats in this section.

The server project of nats is gnatsd. The communication method between client and gnatsd is tcp-based text protocol, which is very simple:

Send a message to the subject for the task:

![nats-protocol-pub](../images/ch6-09-nats-protocol-pub.png)

*Figure 6-15 pub* in the nats protocol

Subscribe to the tasks subject with the queue of workers:

![nats-protocol-sub](../images/ch6-09-nats-protocol-sub.png)

*Figure 6-16 sub* in the nats protocol

The queue parameter is optional. If you want to load balance the task on the distributed consumer side, instead of everyone receiving the same message, you should assign the same queue name to the consumer.

#### Basic message production

Production messages can be specified by specifying the subject:

```go
Nc, err := nats.Connect(nats.DefaultURL)
If err != nil {return}

/ / Specify subject is tasks, the content of the message is free
Err = nc.Publish("tasks", []byte("your task content"))

nc.Flush()
```

#### Basic message consumption

Direct use of the subscription API of Nats does not achieve the purpose of task distribution, because the pub sub itself is broadcast. All consumers will receive exactly the same message.

In addition to the normal subscribe, nats also provides the function of queue subscribe. By providing a queue group name (similar to the consumer group in Kafka), tasks can be distributed to consumers in a balanced manner.

```go
Nc, err := nats.Connect(nats.DefaultURL)
If err != nil {return}

// queue subscribe is equivalent to branch balancing for task distribution between consumers
// The premise is that all consumers use the worker queue
// The queue in nats is conceptually similar to the consumer group in Kafka
Sub, err := nc.QueueSubscribeSync("tasks", "workers")
If err != nil {return}

Var msg *nats.Msg
For {
Msg, err = sub.NextMsg(time.Hour * 10000)
If err != nil {break}
// correctly consumed the message
// can handle tasks with nats.Msg object
}
```

## 6.7.3 Combining the production of messages with nats and colly

We customize a corresponding collector for each website, and set the corresponding rules, such as abcdefg, hijklmn (fictional), and then use a simple factory method to map the collector to its host. Each site climbs to the list page. You need to parse all the links in the current program and send the URL of the landing page to the message queue.

```go
Package main

Import (
"fmt"
"net/url"

"github.com/gocolly/colly"
)

Var domain2Collector = map[string]*colly.Collector{}
Var nc *nats.Conn
Var maxDepth = 10
Var natsURL = "nats://localhost:4222"

Func factory(urlStr string) *colly.Collector {
u, _ := url.Parse(urlStr)
Return domain2Collector[u.Host]
}

Func initABCDECollector() *colly.Collector {
c := colly.NewCollector(
colly.AllowedDomains("www.abcdefg.com"),
colly.MaxDepth(maxDepth),
)

c.OnResponse(func(resp *colly.Response) {
// Do some aftercare work after climbing
// For example, the confirmation that the page has been crawled is stored in MySQL.
})

c.OnHTML("a[href]", func(e *colly.HTMLElement) {
// Basic anti-reptile strategy
Link := e.Attr("href")
time.Sleep(time.Second * 2)

// If you match the list page, you will visit
If listRegex.Match([]byte(link)) {
c.Visit(e.Request.AbsoluteURL(link))
}
// If the regular match is on the landing page, send a message queue.
If detailRegex.Match([]byte(link)) {
Err = nc.Publish("tasks", []byte(link))
nc.Flush()
}
})
Return c
}

Func initHIJKLCollector() *colly.Collector {
c := colly.NewCollector(
colly.AllowedDomains("www.hijklmn.com"),
colly.MaxDepth(maxDepth),
)

c.OnHTML("a[href]", func(e *colly.HTMLElement) {
})

Return c
}

Func init() {
domain2Collector["www.abcdefg.com"] = initABCDECollector()
domain2Collector["www.hijklmn.com"]= initHIJKLCollector()
Var err error
Nc, err = nats.Connect(natsURL)
If err != nil {os.Exit(1)}
}

Func main() {
Urls := []string{"https://www.abcdefg.com", "https://www.hijklmn.com"}
For _, url := range urls {
Instance := factory(url)
instance.Visit(url)
}
}

```

## 6.7.4 Combining message consumption with collly

The consumer side is a bit simpler, we only need to subscribe to the corresponding theme, and directly visit the website's details page (flooring page).

```go
Package main

Import (
"fmt"
"net/url"

"github.com/gocolly/colly"
)

Var domain2Collector = map[string]*colly.Collector{}
Var nc *nats.Conn
Var maxDepth = 10
Var natsURL = "nats://localhost:4222"

Func factory(urlStr string) *colly.Collector {
u, _ := url.Parse(urlStr)
Return domain2Collector[u.Host]
}

Func initV2exCollector() *colly.Collector {
c := colly.NewCollector(
colly.AllowedDomains("www.abcdefg.com"),
colly.MaxDepth(maxDepth),
)
Return c
}

Func initV2fxCollector() *colly.Collector {
c := colly.NewCollector(
colly.AllowedDomains("www.hijklmn.com"),
colly.MaxDepth(maxDepth),
)
Return c
}

Func init() {
domain2Collector["www.abcdefg.com"] = initV2exCollector()
domain2Collector["www.hijklmn.com"] = initV2fxCollector()

Var err error
Nc, err = nats.Connect(natsURL)
If err != nil {os.Exit(1)}
}

Func startConsumer() {
Nc, err := nats.Connect(nats.DefaultURL)
If err != nil {return}

Sub, err := nc.QueueSubscribeSync("tasks", "workers")
If err != nil {return}

Var msg *nats.Msg
For {
Msg, err = sub.NextMsg(time.Hour * 10000)
If err != nil {break}

urlStr := string(msg.Data)
Ins := factory(urlStr)
// Because the most downstream one must be the landing page of the corresponding website.
// So don’t have to make extra judgments, just climb the content directly.
ins.Visit(urlStr)
// prevent being blocked
time.Sleep(time.Second)
}
}

Func main() {
startConsumer()
}
```

At the code level, the producers and consumers here are essentially the same. If we want to flexibly support the increase and decrease of crawling of various websites in the future, we should think about how to configure these crawler strategies and parameters as much as possible.

The use of some configuration systems has been covered in the Distributed Configuration section of this chapter, and readers can try it out on their own, so I won't go into details here.