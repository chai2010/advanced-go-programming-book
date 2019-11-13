# 5.1 Introduction to Web Development

Because Go's `net/http` package provides a basic combination of routing functions and rich functional functions. So in the community, there is a popular idea that writing APIs without Go doesn't require a framework. In our opinion, if your project's routing is in single digits, the URI is fixed, and the parameters are not passed through the URI, then the official library is used. enough. But in complex scenarios, the official http library is still a bit weak. For example, the following route:

```
GET /card/:id
POST /card/:id
DELTE /card/:id
GET /card/:id/name
...
GET /card/:id/relations
```

It can be seen whether the framework is used or whether the specific problem is specifically analyzed.

Go's web framework can be roughly divided into two categories:

1. Router framework
2. MVC class framework

In the choice of the framework, in most cases, according to personal preferences and the company's technology stack. For example, if the company has a lot of technicians who are PHP, then they will definitely like the framework like beego, but if the company has a lot of C programmers, then their ideas may be as simple as possible. For example, many C programmers in large factories may even use C language to write small CGI programs. They may not have the willingness to learn MVC or more complex Web frameworks. All they need is a very simple route. Even the route is not needed, only a basic HTTP protocol processing library is needed to help him save the manual labor.

Go's `net/http` package provides such basic functionality, and writing a simple `http echo server` takes only 30 seconds.

```go
//brief_intro/echo.go
Package main
Import (...)

Func echo(wr http.ResponseWriter, r *http.Request) {
Msg, err := ioutil.ReadAll(r.Body)
If err != nil {
wr.Write([]byte("echo error"))
Return
}

writeLen, err := wr.Write(msg)
If err != nil || writeLen != len(msg) {
log.Println(err, "write len:", writeLen)
}
}

Func main() {
http.HandleFunc("/", echo)
Err := http.ListenAndServe(":8080", nil)
If err != nil {
log.Fatal(err)
}
}

```

If you haven't finished the program after 30s, please check if your own typing speed is slow (Just kidding :D). This example is to illustrate how easy it is to write a small program for the HTTP protocol in Go. If you are faced with a more complex situation, such as enterprise applications with dozens of interfaces, it is not appropriate to use the `net/http` library directly.

Let's take a look at the practices in a Kafka monitoring project in the open source community:

```go
//Burrow: http_server.go
Func NewHttpServer(app *ApplicationContext) (*HttpServer, error) {
...
server.mux.HandleFunc("/", handleDefault)

server.mux.HandleFunc("/burrow/admin", handleAdmin)

server.mux.Handle("/v2/kafka", appHandler{server.app, handleClusterList})
server.mux.Handle("/v2/kafka/", appHandler{server.app, handleKafka})
server.mux.Handle("/v2/zookeeper", appHandler{server.app, handleClusterList})
...
}
```

The above code comes from Burrow, the Kafka monitoring project of the famous linkedin company. It does not use any router framework and only uses `net/http`. Just looking at the code above seems very elegant, there are only five simple URIs in our project, so the service we provide is like this:

```go
/
/burrow/admin
/v2/kafka
/v2/kafka/
/v2/zookeeper
```

If you really think so, you are cheated. Let's take a look at the `handleKafka()` function to find out:

```go
Func handleKafka(app *ApplicationContext, w http.ResponseWriter, r *http.Request) (int, string) {
pathParts := strings.Split(r.URL.Path[1:], "/")
If _, ok := app.Config.Kafka[pathParts[2]]; !ok {
Return makeErrorResponse(http.StatusNotFound, "cluster not found", w, r)
}
If pathParts[2] == "" {
// Allow a trailing / on requests
Return handleClusterList(app, w, r)
}
If (len(pathParts) == 3) || (pathParts[3] == "") {
Return handleClusterDetail(app, w, r, pathParts[2])
}

Switch pathParts[3] {
Case "consumer":
Switch {
Case r.Method == "DELETE":
Switch {
Case (len(pathParts) == 5) || (pathParts[5] == ""):
Return handleConsumerDrop(app, w, r, pathParts[2], pathParts[4])
Default:
Return makeErrorResponse(http.StatusMethodNotAllowed, "request method not supported", w, r)
}
Case r.Method == "GET":
Switch {
Case (len(pathParts) == 4) || (pathParts[4] == ""):
Return handleConsumerList(app, w, r, pathParts[2])
Case (len(pathParts) == 5) || (pathParts[5] == ""):
// Consumer detail - list of consumer streams/hosts? Can be config info later
Return makeErrorResponse(http.StatusNotFound, "unknown API call", w, r)
Case pathParts[5] == "topic":
Switch {
Case (len(pathParts) == 6) || (pathParts[6] == ""):
Return handleConsumerTopicList(app, w, r, pathParts[2], pathParts[4])
Case (len(pathParts) == 7) || (pathParts[7] == ""):
Return handleConsumerTopicDetail(app, w, r, pathParts[2], pathParts[4], pathParts[6])
}
Case pathParts[5] == "status":
Return handleConsumerStatus(app, w, r, pathParts[2], pathParts[4], false)
Case pathParts[5] == "lag":
Return handleConsumerStatus(app, w, r, pathParts[2], pathParts[4], true)
}
Default:
Return makeErrorResponse(http.StatusMethodNotAllowed, "request method not supported", w, r)
}
Case "topic":
Switch {
Case r.Method != "GET":
Return makeErrorResponse(http.StatusMethodNotAllowed, "request method not supported", w, r)
Case (len(pathParts) == 4) || (pathParts[4] == ""):
Return handleBrokerTopicList(app, w, r, pathParts[2])
Case (len(pathParts) == 5) || (pathParts[5] == ""):
Return handleBrokerTopicDetail(app, w, r, pathParts[2], pathParts[4])
}
Case "offsets":
// Reserving this endpoint to implement later
Return makeErrorResponse(http.StatusNotFound, "unknown API call", w, r)
}

// If we fell through, return a 404
Return makeErrorResponse(http.StatusNotFound, "unknown API call", w, r)
}
```

Because `mux` in the default `net/http` package does not support routing with parameters, Burrow uses a very crappy string `Split` and a messy `switch case` to achieve its goal, but The routing management logic that should be concentrated is complicated, scattered throughout the system, and difficult to maintain and manage. If you look at the code carefully, you may find that several other `handler` functions are logically simpler, and the most complicated one is the `handleKafka()`. And our system always starts to grow up from such insignificant confusion, and eventually becomes difficult to clean up.

In our experience, simply, as long as your route has parameters and the number of APIs for this project exceeds 10, try not to use the default route in `net/http`. The most widely used router in Go open source is httpRouter, and many open source router frameworks are based on httpRouter to achieve a certain degree of transformation. The principle of httpRouter routing is explained in detail in the router section of this chapter.

Looking back at the beginning of the article, there are several frameworks in the open source world. The first one is to simply wrap httpRouter and then provide custom middleware and some simple gadget integration such as gin, which is lightweight, easy to learn, and high. performance. The second is to learn from some other language programming styles of some MVC class frameworks, such as beego, to facilitate the rapid migration of programmers from other languages. There are also some more powerful frameworks, except for the database schema. Design, most of the code is generated directly, such as goa. Regardless of the framework, the best for the developer's background.

In addition to the principles of the router and middleware, the contents of this chapter will be combined with Go to make some practical explanations. I hope to be able to help readers who have not been exposed to relevant content.