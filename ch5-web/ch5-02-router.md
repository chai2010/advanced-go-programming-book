# 5.2 router request routing

In the common web framework, the router is a must-have component. The router in the Go language circle is often referred to as the `http` multiplexer. In the previous section, we learned how to use the mux built into the `http` standard library to perform simple routing functions through simple learning of Burrow code. If the development Web system is not interested in the parameters in the path, use `mux` in the `http` standard library.

RESTful is a wave of API design that started a few years ago. In addition to GET and POST, RESTful uses several other standardized semantics defined by the HTTP protocol. Specifically include:

```go
Const (
MethodGet = "GET"
MethodHead = "HEAD"
MethodPost = "POST"
MethodPut = "PUT"
MethodPatch = "PATCH" // RFC 5789
MethodDelete = "DELETE"
MethodConnect = "CONNECT"
MethodOptions = "OPTIONS"
MethodTrace = "TRACE"
)
```

Take a look at the common request paths in RESTful:

```shell
GET /repos/:owner/:repo/comments/:id/reactions

POST /projects/:project_id/columns

PUT /user/starred/:owner/:repo

DELETE /user/starred/:owner/:repo
```

Believe that you are smart, you have already guessed it. This is a few API designs selected in Github's official documentation. The RESTful style API relies heavily on the request path. Many parameters are placed in the request URI. In addition to this, many less common HTTP status codes are used, but this section only discusses routing, so skip it first.

If our system also wants such a URI design, the `mux` using the standard library is obviously not enough.

## 5.2.1 httprouter

The more popular open source go Web frameworks mostly use httprouter, or support for routing based on variants of httprouter. The parametric routing of github mentioned above can be supported in httprouter.

Because the explicit use of httprouter is used, you need to avoid some situations that lead to routing conflicts when designing routes, for example:

```
Conflict:
GET /user/info/:name
GET /user/:id

No conflict:
GET /user/info/:name
POST /user/:id
```

In a nutshell, if two routes have a consistent http method (referred to as GET/POST/PUT/DELETE) and a request path prefix, and an A route appears at a certain location, it is a wildcard (referred to as: id) parameter. B route is a normal string, then a route conflict will occur. Routing conflicts will panic directly during the initialization phase:

```shell
Panic: wildcard route ':id' conflicts with existing children in path '/user/:id'

Goroutine 1 [running]:
Github.com/cch123/httprouter.(*node).insertChild(0xc4200801e0, 0xc42004fc01, 0x126b177, 0x3, 0x126b171, 0x9, 0x127b668)
  /Users/caochunhui/go_work/src/github.com/cch123/httprouter/tree.go:256 +0x841
Github.com/cch123/httprouter.(*node).addRoute(0xc4200801e0, 0x126b171, 0x9, 0x127b668)
  /Users/caochunhui/go_work/src/github.com/cch123/httprouter/tree.go:221 +0x22a
Github.com/cch123/httprouter.(*Router).Handle(0xc42004ff38, 0x126a39b, 0x3, 0x126b171, 0x9, 0x127b668)
  /Users/caochunhui/go_work/src/github.com/cch123/httprouter/router.go:262 +0xc3
Github.com/cch123/httprouter.(*Router).GET(0xc42004ff38, 0x126b171, 0x9, 0x127b668)
  /Users/caochunhui/go_work/src/github.com/cch123/httprouter/router.go:193 +0x5e
Main.main()
  /Users/caochunhui/test/go_web/httprouter_learn2.go:18 +0xaf
Exit status 2
```

Another point to note is that since httprouter takes into account the depth of the dictionary tree, the number of parameters is limited during initialization, so the number of parameters in the route cannot exceed 255. Otherwise, httprouter will not recognize subsequent parameters. However, there is no need to think too much on this point. After all, URI is designed and given to people. I believe that no exaggerated URI can have more than 200 parameters in a path.

In addition to the wildcard parameter in the support path, httprouter can also support the `*` number for wildcarding, but the parameter at the beginning of `*` can only be placed at the end of the route, for example:

```shell
Pattern: /src/*filepath

 /src/ filepath = ""
 /src/somefile.go filepath = "somefile.go"
 /src/subdir/somefile.go filepath = "subdir/somefile.go"
```

This design may be less common in RESTful, mainly to enable a simple HTTP static file server using httprouter.

In addition to the normal routing support, httprouter also supports customization of callback functions in special cases, such as 404:

```go
r := httprouter.New()
r.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("oh no, not found"))
})
```

Or when internal panic:
```go
r.PanicHandler = func(w http.ResponseWriter, r *http.Request, c interface{}) {
log.Printf("Recovering from panic, Reason: %#v", c.(error))
w.WriteHeader(http.StatusInternalServerError)
w.Write([]byte(c.(error).Error()))
}
```

Currently the most popular (star-numbered) web framework [gin] (https://github.com/gin-gonic/gin) in the open source world is a variant of httprouter.

## 5.2.2 Principle

The data structure used by httprouter and many derived routers is called the Radix Tree. The reader may not have been exposed to the compressed dictionary tree, but should have heard about the dictionary tree (Trie Tree). * Figure 5-1* is a typical dictionary tree structure:

![trie tree](../images/ch6-02-trie.png)

*Figure 5-1 Dictionary Tree*

The dictionary tree is often used for string retrieval, such as building a dictionary tree with a given sequence of strings. For the target string, as long as the depth-first search is started from the root node, it can be judged whether the string has ever appeared, and the time complexity is `O(n)`, and n can be regarded as the length of the target string. Why do you want to do this? Strings themselves are not numerically comparable to numeric types, and the time complexity of two string comparisons depends on the length of the string. If you do not use the dictionary tree to complete the above functions, you need to sort the history strings, and then use the algorithm such as binary search to search, the time complexity is only high. The dictionary tree can be considered as a typical way of space-changing time.

The common dictionary tree has a obvious disadvantage, that is, each letter needs to establish a child node, which will lead to a deeper dictionary tree, and the compression dictionary tree balances the advantages and disadvantages of the dictionary tree relatively well. Is a typical compression dictionary tree structure:

![radix tree](../images/ch6-02-radix.png)

*Figure 5-2 Compressed dictionary tree*

Not only one letter is stored on each node, which is also the main meaning of "compression" in the compressed dictionary tree. Using a compressed dictionary tree can reduce the number of layers in the tree, and because the data storage on each node is more than the usual dictionary tree, the locality of the program is better (a node's path can be loaded into the cache to perform multiple characters. Contrast), thus making the CPU cache friendly.

## 5.2.3 Compressed dictionary tree creation process

Let's track the creation process of a typical compressed dictionary tree in httprouter. The routing settings are as follows:

```
PUT /user/installations/:installation_id/repositories/:repository_id

GET /marketplace_listing/plans/
GET /marketplace_listing/plans/:id/accounts
GET /search
GET /status
GET /support

Supplementary route:
GET /marketplace_listing/plans/ohyes
```

The last supplementary route is what we imagined, except that all API routes are from `api.github.com`.

### 5.2.3.1 Root Node Creation

The compression dictionary tree stored in the Router structure of httprouter uses the following data structure:

```go
// omitted the other part of the Router struct
Type Router struct {
// ...
Trees map[string]*node
// ...
}
```

The `key` in `trees` is the various methods defined in the RFC of HTTP 1.1, specifically:

```shell
GET
HEAD
OPTIONS
POST
PUT
PATCH
DELETE
```

Each method corresponds to an independent compressed dictionary tree that does not share data with each other. Specific to the route we used above, `PUT` and `GET` are two trees instead of one.

Simply put, the first time a method inserts a route, the root node of the corresponding dictionary tree is created. In order, we first have a `PUT`:

```go
r := httprouter.New()
r.PUT("/user/installations/:installation_id/repositories/:reposit", Hello)
```

Thus the root node corresponding to `PUT` will be created. Draw this tree of `PUT`:

![put radix tree](../images/ch6-02-radix-put.png)

*Figure 5-3 Compressed dictionary tree after inserting a route*

The node type of radix is ​​`*httprouter.node`. For the convenience of explanation, we have left a few fields that are currently concerned:

```
Path: the string in the path corresponding to the current node

wildChild: whether the child node is a parameter node, ie wildcard node, or :id

nType: current node type, with four enumeration values: static/root/param/catchAll.
    Static // ordinary string node of the non-root node
    Root // root node
    Param // parameter node, for example: id
    catchAll // wildcard node, for example *anyway

Indices: child node index, when the child node is non-parameter type, that is, the wildchild of the node is false, the first letter of each child node is placed in the index array. Said to be an array, actually a string.

```

Of course, the `PUT` route has only one path. Next, we take the following multiple GET paths as an example to explain the insertion process of the child nodes.

### 5.2.3.2 Subnode insertion

When inserting `GET /marketplace_listing/plans`, similar to the previous PUT process, the structure of the GET tree is as shown in Figure 5-4*:

![get radix step 1](../images/ch6-02-radix-get-1.png)

*Figure 5-4 Inserting the compression dictionary tree of the first node*

Because the first route has no parameters, the path is stored on the root node. So there is only one node.

Then insert `GET /marketplace_listing/plans/:id/accounts`, the new path has the same prefix as the previous path, and can be inserted directly after the previous leaf node, then the result is very simple, see the inserted tree structure *Figure 5-5*:

![get radix step 2](../images/ch6-02-radix-get-2.png)

*Figure 5-5 Inserting the compressed dictionary tree of the second node*

Since the `:id` node has only a normal child of a string, the indices still do not need to be processed.

The above situation is relatively simple, and the new route can be directly inserted as a child node of the original route. The actual situation will not be so good.

### 5.2.3.3 Edge splitting

Next we insert `GET /search`, which will cause the edge of the tree to split, see *Figure 5-6*.

![get radix step 3](../images/ch6-02-radix-get-3.png)

*Figure 5-6 Inserting the third node causes the edge to split*

The original path and the new path are split at the initial `/` position, so that the original root node content needs to be moved down, and the new route `search` is also suspended as a child node under the root node. At this time, because there are multiple child nodes, the root node's indices provide the child node index, and this field needs to come in handy. "ms" represents the initials of the child nodes as m (marketplace) and s (search).

We sighed and inserted `GET /status` and `GET /support` into the tree. This will cause a split on the `search` node again. The final result is shown in Figure 5-7*:

![get radix step 4](../images/ch6-02-radix-get-4.png)

*Figure 5-7 Compressed dictionary tree after inserting all routes*

### 5.2.3.4 Subnode conflict handling

In the case where the route itself has only strings, no conflicts will occur. Conflicts may only occur if the route contains wildcards (like :id) or catchAll. This has already been mentioned before.

The conflict handling of child nodes is very simple, in several cases:

1. When inserting a wildcard node, the parent node's children array is not empty and wildChild is set to false. For example: `GET /user/getAll` and `GET /user/:id/getAddr`, or `GET /user/*aaa` and `GET /user/:id`.
2. When inserting a wildcard node, the parent node's children array is not empty and wildChild is set to true, but the parent card's wildcard child node has a different wildcard name to insert. For example: `GET /user/:id/info` and `GET /user/:name/info`.
3. When inserting the catchAll node, the children of the parent node are not empty. For example: `GET /src/abc` and `GET /src/*filename`, or `GET /src/:id` and `GET /src/*filename`.
4. When the static node is inserted, the wildChild field of the parent node is set to true.
5. When inserting a static node, the children of the parent node are not empty, and the child node nType is catchAll.

As long as a conflict occurs, it will panic at initialization time. For example, when inserting our desired route `GET /marketplace_listing/plans/ohyes`, a fourth conflict situation occurs: the wildChild field of its parent `marketplace_listing/plans/` is true.