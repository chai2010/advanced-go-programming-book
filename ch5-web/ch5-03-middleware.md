# 5.3 Middleware

This chapter will analyze the middleware technology principles in today's popular web framework and show how to use middleware technology to decouple business and non-business code functions.

## 5.3.1 Code mire

Let's look at a piece of code:

```go
// middleware/hello.go
Package main

Func hello(wr http.ResponseWriter, r *http.Request) {
wr.Write([]byte("hello"))
}

Func main() {
http.HandleFunc("/", hello)
Err := http.ListenAndServe(":8080", nil)
...
}
```

This is a typical web service that mounts a simple route. Our online services are generally gradually expanding from such a simple service.

Now suddenly there is a new demand, we want to count the processing time of the hello service written before, the demand is very simple, we make a few modifications to the above program:

```go
// middleware/hello_with_time_elapse.go
Var logger = log.New(os.Stdout, "", 0)

Func hello(wr http.ResponseWriter, r *http.Request) {
timeStart := time.Now()
wr.Write([]byte("hello"))
timeElapsed := time.Since(timeStart)
logger.Println(timeElapsed)
}
```

This allows you to print out the time spent on the current request each time you receive an http request.

After completing this requirement, we continued our business development, and the APIs provided gradually increased. Now our route looks like this:

```go
// middleware/hello_with_more_routes.go
// omitted some of the same code
Package main

Func helloHandler(wr http.ResponseWriter, r *http.Request) {
// ...
}

Func showInfoHandler(wr http.ResponseWriter, r *http.Request) {
// ...
}

Func showEmailHandler(wr http.ResponseWriter, r *http.Request) {
// ...
}

Func showFriendsHandler(wr http.ResponseWriter, r *http.Request) {
timeStart := time.Now()
wr.Write([]byte("your friends is tom and alex"))
timeElapsed := time.Since(timeStart)
logger.Println(timeElapsed)
}

Func main() {
http.HandleFunc("/", helloHandler)
http.HandleFunc("/info/show", showInfoHandler)
http.HandleFunc("/email/show", showEmailHandler)
http.HandleFunc("/friends/show", showFriendsHandler)
// ...
}

```

Each handler has the code for the record runtime mentioned earlier. Every time we add a new route, we also need to copy these seemingly similar code to the place we need. Because the code is not too much, there is no big problem in implementation.

Gradually our system has been added to 30 routes and `handler` functions. Each time we add a new handler, our first job is to copy all the previously written peripheral code that is not related to the business logic.

Then the system runs safely for a period of time. Suddenly one day, the boss finds you. We recently found someone to develop a new monitoring system. In order to make the system run more controllable, we need to report the time-consuming data of each interface to us. In the monitoring system. Give the monitoring system a name, called metrics. Now you need to modify the code and send the time to the metrics system via HTTP Post. Let's modify `helloHandler()`:

```go
Func helloHandler(wr http.ResponseWriter, r *http.Request) {
timeStart := time.Now()
wr.Write([]byte("hello"))
timeElapsed := time.Since(timeStart)
logger.Println(timeElapsed)
// Add time-consuming reporting
metrics.Upload("timeHandler", timeElapsed)
}
```

Modified here, instinctively found that our development work began to fall into a quagmire. Regardless of any other non-functional or statistical needs of our Web system in the future, our modifications will inevitably lead to a whole body. As long as we add a very simple non-business statistic, we need to add dozens of handlers to add these business-independent code. Although we didn't seem to have made a mistake at the beginning, it was clear that as the business evolved, our way of doing things caught us in the quagmire of code.

## 5.3.2 Using middleware to strip non-business logic

Let's analyze it. Where did it go wrong at the beginning? We just meet the demand step by step, write down the logic we need according to the process?

The biggest mistake we made was to put business code and non-business code together. For most scenarios, the non-business requirement is to do something before the http request is processed, and do something after the response is complete. Can we use some refactoring ideas to strip out these public non-business function code? Going back to the example at the beginning, we need to add timeout statistics to our `helloHandler()`. We can wrap `helloHandler()` with a method called `function adapter`:

```go

Func hello(wr http.ResponseWriter, r *http.Request) {
wr.Write([]byte("hello"))
}

Func timeMiddleware(next http.Handler) http.Handler {
Return http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) {
timeStart := time.Now()

// next handler
next.ServeHTTP(wr, r)

timeElapsed := time.Since(timeStart)
logger.Println(timeElapsed)
})
}

Func main() {
http.Handle("/", timeMiddleware(http.HandlerFunc(hello)))
Err := http.ListenAndServe(":8080", nil)
...
}
```

This makes it very easy to achieve a strip between business and non-business, the magic lies in this `timeMiddleware`. As you can see from the code, our `timeMiddleware()` is also a function whose parameters are `http.Handler`, and the definition of `http.Handler` is in the `net/http` package:

```go
Type Handler interface {
ServeHTTP(ResponseWriter, *Request)
}
```

Any method implements `ServeHTTP`, which is a legal `http.Handler`. You may have some confusion when reading this. Let's sort out the relationship between `Handler`, `HandlerFunc` and `ServeHTTP` of http library. :

```go
Type Handler interface {
ServeHTTP(ResponseWriter, *Request)
}

Type HandlerFunc func(ResponseWriter, *Request)

Func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
f(w, r)
}

```
As long as your handler function signature is:

```go
Func (ResponseWriter, *Request)
```

Then the `handler` and `http.HandlerFunc()` have a consistent function signature, and the `handler()` function can be type converted to `http.HandlerFunc`. And `http.HandlerFunc` implements the `http.Handler` interface. When the `http` library needs to call your handler function to process the http request, it will call the `ServeHTTP()` function of `HandlerFunc()`. The basic call chain of a request is like this:

```go
h = getHandler() => h.ServeHTTP(w, r) => h(w, r)
```

The above mentioned conversion of the custom `handler` to `http.HandlerFunc()` is necessary because our `handler` does not directly implement the `ServeHTTP` interface. The CastleFunc (note the difference between HandlerFunc and HandleFunc) we saw in the above code can also see this cast process:

```go
Func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
DefaultServeMux.HandleFunc(pattern, handler)
}

// transfer

Func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
mux.Handle(pattern, HandlerFunc(handler))
}
```

Knowing what the handler is all about, our middleware understands by wrapping the handler and returning a new handler.

To summarize, what our middleware does is wrap the handler through one or more functions, returning a function chain that includes the logic of each middleware. We have made the above packaging more complicated:

```go
customizedHandler = logger(timeout(ratelimit(helloHandler)))
```

The context of this function chain during execution can be represented by Figure 5-8*.

![](../images/ch6-03-middleware_flow.png)

*Figure 5-8 Request Processing *

To be more straightforward, this process is constantly pushing the function and then popping it out when the request is processed. There are some execution flows similar to recursion:

```
[exec of logger logic] function stack: []

[exec of timeout logic] Function stack: [logger]

[exec of ratelimit logic] Function stack: [timeout/logger]

[exec of helloHandler logic] Function stack: [ratelimit/timeout/logger]

[exec of ratelimit logic part2] Function stack: [timeout/logger]

[exec of timeout logic part2] Function stack: [logger]

[exec of logGer logic part2] function stack: []
```

The function is implemented, but we also saw in the above use, the use of this function set function is not very beautiful, and it does not have any readability.

## 5.3.3 More elegant middleware

In the previous section, the decoupling of business function code and non-business function code was solved, but it also mentioned that it does not look good. If you need to modify the order of these functions, or add or delete middleware is still a little hard, this section we will carry out Some optimizations on "writing".

See an example:

```go
r = NewRouter()
r.Use(logger)
r.Use(timeout)
r.Use(ratelimit)
r.Add("/", helloHandler)
```

With multi-step setup, we have a chain of execution functions similar to the previous one. Winning is intuitive and easy to understand. If we want to add or remove middleware, simply add the corresponding `Use()` call to delete it. Very convenient.

From the perspective of the framework, how to achieve such a function? Not complicated:

```go
Type middleware func(http.Handler) http.Handler

Type Router struct {
middlewareChain [] middleware
Mux map[string] http.Handler
}

Func NewRouter() *Router{
Return &Router{}
}

Func (r *Router) Use(m middleware) {
r.middlewareChain = append(r.middlewareChain, m)
}

Func (r *Router) Add(route string, h http.Handler) {
Var mergedHandler = h

For i := len(r.middlewareChain) - 1; i >= 0; i-- {
mergedHandler = r.middlewareChain[i](mergedHandler)
}

R.mux[route] = mergedHandler
}
```

Note that the `middleware` array traversal order in the code should be "opposite" to the order the user wants to call. It should not be difficult to understand.


## 5.3.4 What is suitable for doing in middleware?

Take the more popular open source Go language framework as an example:

```
Compress.go
  => Compress the response body of http
Heartbeat.go
  => Set a special route, such as /ping, /healthcheck, to probe the pre-services such as load balancing
Logger.go
  => Print request processing log, such as request processing time, request routing
Profiler.go
  => Mount the route required by pprof, such as `/pprof`, `/pprof/trace` to the system
Realip.go
  => Read X-Forwarded-For and X-Real-IP from the request header, and modify RemoteAddr in http.Request to get RealIP
Requestid.go
  => Generate a separate requestid for this request, which can be used to generate distributed call links, or all logic for concatenating single requests in the log.
Timeout.go
  => Set the timeout with context.Timeout and pass it through http.Request
Throttler.go
  => Store tokens through fixed-length channels and limit the interface through these tokens
```

Each web framework will have a corresponding middleware component. If you are interested, you can also contribute useful middleware to these projects. As long as the reasonable general project maintainers are willing to merge your Pull Request.

For example, the framework of gin, which is very popular in the open source world, has opened a warehouse for middleware contributed by users. See Figure 5-9*:

![](../images/ch6-03-gin_contrib.png)

*Figure 5-9 gin middleware repository*

If the reader reads the source code of gin, you may find that the gin middleware does not handle `http.Handler`, but a function type called `gin.HandlerFunc`, and the `http.Handler explained in this section. `Signature is not the same. However, gin's `handler` is just a package for its framework. The principle of middleware is consistent with the description in this section.