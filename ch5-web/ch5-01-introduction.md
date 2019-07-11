# 5.1 Web 开发简介

因为Go的`net/http`包提供了基础的路由函数组合与丰富的功能函数。所以在社区里流行一种用Go编写API不需要框架的观点，在我们看来，如果你的项目的路由在个位数、URI固定且不通过URI来传递参数，那么确实使用官方库也就足够。但在复杂场景下，官方的http库还是有些力有不逮。例如下面这样的路由：

```
GET   /card/:id
POST  /card/:id
DELTE /card/:id
GET   /card/:id/name
...
GET   /card/:id/relations
```

可见是否使用框架还是要具体问题具体分析的。

Go的Web框架大致可以分为这么两类：

1. Router框架
2. MVC类框架

在框架的选择上，大多数情况下都是依照个人的喜好和公司的技术栈。例如公司有很多技术人员是PHP出身，那么他们一定会非常喜欢像beego这样的框架，但如果公司有很多C程序员，那么他们的想法可能是越简单越好。比如很多大厂的C程序员甚至可能都会去用C语言去写很小的CGI程序，他们可能本身并没有什么意愿去学习MVC或者更复杂的Web框架，他们需要的只是一个非常简单的路由（甚至连路由都不需要，只需要一个基础的HTTP协议处理库来帮他省掉没什么意思的体力劳动）。

Go的`net/http`包提供的就是这样的基础功能，写一个简单的`http echo server`只需要30s。

```go
//brief_intro/echo.go
package main
import (...)

func echo(wr http.ResponseWriter, r *http.Request) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		wr.Write([]byte("echo error"))
		return
	}

	writeLen, err := wr.Write(msg)
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
}

func main() {
	http.HandleFunc("/", echo)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

```

如果你过了30s还没有完成这个程序，请检查一下你自己的打字速度是不是慢了（开个玩笑 :D）。这个例子是为了说明在Go中写一个HTTP协议的小程序有多么简单。如果你面临的情况比较复杂，例如几十个接口的企业级应用，直接用`net/http`库就显得不太合适了。

我们来看看开源社区中一个Kafka监控项目中的做法：

```go
//Burrow: http_server.go
func NewHttpServer(app *ApplicationContext) (*HttpServer, error) {
	...
	server.mux.HandleFunc("/", handleDefault)

	server.mux.HandleFunc("/burrow/admin", handleAdmin)

	server.mux.Handle("/v2/kafka", appHandler{server.app, handleClusterList})
	server.mux.Handle("/v2/kafka/", appHandler{server.app, handleKafka})
	server.mux.Handle("/v2/zookeeper", appHandler{server.app, handleClusterList})
	...
}
```

上面这段代码来自大名鼎鼎的linkedin公司的Kafka监控项目Burrow，没有使用任何router框架，只使用了`net/http`。只看上面这段代码似乎非常优雅，我们的项目里大概只有这五个简单的URI，所以我们提供的服务就是下面这个样子：

```go
/
/burrow/admin
/v2/kafka
/v2/kafka/
/v2/zookeeper
```

如果你确实这么想的话就被骗了。我们再进`handleKafka()`这个函数一探究竟：

```go
func handleKafka(app *ApplicationContext, w http.ResponseWriter, r *http.Request) (int, string) {
	pathParts := strings.Split(r.URL.Path[1:], "/")
	if _, ok := app.Config.Kafka[pathParts[2]]; !ok {
		return makeErrorResponse(http.StatusNotFound, "cluster not found", w, r)
	}
	if pathParts[2] == "" {
		// Allow a trailing / on requests
		return handleClusterList(app, w, r)
	}
	if (len(pathParts) == 3) || (pathParts[3] == "") {
		return handleClusterDetail(app, w, r, pathParts[2])
	}

	switch pathParts[3] {
	case "consumer":
		switch {
		case r.Method == "DELETE":
			switch {
			case (len(pathParts) == 5) || (pathParts[5] == ""):
				return handleConsumerDrop(app, w, r, pathParts[2], pathParts[4])
			default:
				return makeErrorResponse(http.StatusMethodNotAllowed, "request method not supported", w, r)
			}
		case r.Method == "GET":
			switch {
			case (len(pathParts) == 4) || (pathParts[4] == ""):
				return handleConsumerList(app, w, r, pathParts[2])
			case (len(pathParts) == 5) || (pathParts[5] == ""):
				// Consumer detail - list of consumer streams/hosts? Can be config info later
				return makeErrorResponse(http.StatusNotFound, "unknown API call", w, r)
			case pathParts[5] == "topic":
				switch {
				case (len(pathParts) == 6) || (pathParts[6] == ""):
					return handleConsumerTopicList(app, w, r, pathParts[2], pathParts[4])
				case (len(pathParts) == 7) || (pathParts[7] == ""):
					return handleConsumerTopicDetail(app, w, r, pathParts[2], pathParts[4], pathParts[6])
				}
			case pathParts[5] == "status":
				return handleConsumerStatus(app, w, r, pathParts[2], pathParts[4], false)
			case pathParts[5] == "lag":
				return handleConsumerStatus(app, w, r, pathParts[2], pathParts[4], true)
			}
		default:
			return makeErrorResponse(http.StatusMethodNotAllowed, "request method not supported", w, r)
		}
	case "topic":
		switch {
		case r.Method != "GET":
			return makeErrorResponse(http.StatusMethodNotAllowed, "request method not supported", w, r)
		case (len(pathParts) == 4) || (pathParts[4] == ""):
			return handleBrokerTopicList(app, w, r, pathParts[2])
		case (len(pathParts) == 5) || (pathParts[5] == ""):
			return handleBrokerTopicDetail(app, w, r, pathParts[2], pathParts[4])
		}
	case "offsets":
		// Reserving this endpoint to implement later
		return makeErrorResponse(http.StatusNotFound, "unknown API call", w, r)
	}

	// If we fell through, return a 404
	return makeErrorResponse(http.StatusNotFound, "unknown API call", w, r)
}
```

因为默认的`net/http`包中的`mux`不支持带参数的路由，所以Burrow这个项目使用了非常蹩脚的字符串`Split`和乱七八糟的 `switch case`来达到自己的目的，但却让本来应该很集中的路由管理逻辑变得复杂，散落在系统的各处，难以维护和管理。如果读者细心地看过这些代码之后，可能会发现其它的几个`handler`函数逻辑上较简单，最复杂的也就是这个`handleKafka()`。而我们的系统总是从这样微不足道的混乱开始积少成多，最终变得难以收拾。

根据我们的经验，简单地来说，只要你的路由带有参数，并且这个项目的API数目超过了10，就尽量不要使用`net/http`中默认的路由。在Go开源界应用最广泛的router是httpRouter，很多开源的router框架都是基于httpRouter进行一定程度的改造的成果。关于httpRouter路由的原理，会在本章节的router一节中进行详细的阐释。

再来回顾一下文章开头说的，开源界有这么几种框架，第一种是对httpRouter进行简单的封装，然后提供定制的中间件和一些简单的小工具集成比如gin，主打轻量，易学，高性能。第二种是借鉴其它语言的编程风格的一些MVC类框架，例如beego，方便从其它语言迁移过来的程序员快速上手，快速开发。还有一些框架功能更为强大，除了数据库schema设计，大部分代码直接生成，例如goa。不管哪种框架，适合开发者背景的就是最好的。

本章的内容除了会展开讲解router和中间件的原理外，还会以现在工程界面临的问题结合Go来进行一些实践性的说明。希望能够对没有接触过相关内容的读者有所帮助。

