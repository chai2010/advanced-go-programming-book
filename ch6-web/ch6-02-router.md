# 6.2. router 请求路由

在常见的 web 框架中，router 是必备的组件。golang 圈子里 router 也时常被称为 http 的 multiplexer。在上一节中我们通过对 Burrow 代码的简单学习，已经知道如何用 http 标准库中内置的 mux 来完成简单的路由功能了。如果开发 web 系统对路径中带参数没什么兴趣的话，用 http 标准库中的 mux 就可以。

restful 是几年前刮起的 API 设计风潮，在 restful 中使用了 http 标准库还没有支持的一些语义。来看看 restful 中常见的请求路径：

```
GET /repos/:owner/:repo/comments/:id/reactions

POST /projects/:project_id/columns

PUT /user/starred/:owner/:repo

DELETE /user/starred/:owner/:repo
```

相信聪明的你已经猜出来了，这是 github 官方文档中挑出来的几个 api 设计。restful 风格的 API 重度依赖请求路径。会将很多参数放在请求 URI 中。除此之外还会使用很多并不那么常见的 HTTP 状态码，不过本节只讨论路由，所以先略过不谈。

如果我们的系统也想要这样的 URI 设计，使用标准库的 mux 显然就力不从心了。

## httprouter
较流行的开源 golang web 框架大多使用 httprouter，或是基于 httprouter 的变种对路由进行支持。前面提到的 github 的参数式路由在 httprouter 中都是可以支持的。

因为 httprouter 中使用的是显式匹配，所以在设计路由的时候需要规避一些会导致路由冲突的情况，例如：

```
conflict:
GET /user/info/:name
GET /user/:id

no conflict:
GET /user/info/:name
POST /user/:id
```

简单来讲的话，如果两个路由拥有一致的 http method (指 GET/POST/PUT/DELETE) 和请求路径前缀，且在某个位置出现了 A 路由是 wildcard (指 :id 这种形式) 参数，B 路由则是普通字符串，那么就会发生路由冲突。路由冲突会在初始化阶段直接 panic：

```shell
panic: wildcard route ':id' conflicts with existing children in path '/user/:id'

goroutine 1 [running]:
github.com/cch123/httprouter.(*node).insertChild(0xc4200801e0, 0xc42004fc01, 0x126b177, 0x3, 0x126b171, 0x9, 0x127b668)
	/Users/caochunhui/go_work/src/github.com/cch123/httprouter/tree.go:256 +0x841
github.com/cch123/httprouter.(*node).addRoute(0xc4200801e0, 0x126b171, 0x9, 0x127b668)
	/Users/caochunhui/go_work/src/github.com/cch123/httprouter/tree.go:221 +0x22a
github.com/cch123/httprouter.(*Router).Handle(0xc42004ff38, 0x126a39b, 0x3, 0x126b171, 0x9, 0x127b668)
	/Users/caochunhui/go_work/src/github.com/cch123/httprouter/router.go:262 +0xc3
github.com/cch123/httprouter.(*Router).GET(0xc42004ff38, 0x126b171, 0x9, 0x127b668)
	/Users/caochunhui/go_work/src/github.com/cch123/httprouter/router.go:193 +0x5e
main.main()
	/Users/caochunhui/test/go_web/httprouter_learn2.go:18 +0xaf
exit status 2
```

还有一点需要注意，因为 httprouter 考虑到字典树的深度，在初始化时会对参数的数量进行限制，所以在路由中的参数数目不能超过 255，否则会导致 httprouter 无法识别后续的参数。不过这一点上也不用考虑太多，毕竟 URI 是人设计且给人来看的，相信没有变态的 URI 能在一条路径中带有 200 个以上的参数。

除支持路径中的 wildcard 参数之外，httprouter 还可以支持 `*` 号来进行通配，不过 `*` 号开头的参数只能放在路由的结尾，例如下面这样：

```
Pattern: /src/*filepath

 /src/                     filepath = ""
 /src/somefile.go          filepath = "somefile.go"
 /src/subdir/somefile.go   filepath = "subdir/somefile.go"
```

这种设计在 restful 中可能不太常见，主要是为了能够使用 httprouter 来做简单的 http 静态文件服务器。

除了正常情况下的路由支持，httprouter 也支持对一些特殊情况下的回调函数进行定制，例如 404 的时候：

```go
r := httprouter.New()
r.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("oh no, not found"))
})
```

或者内部 panic 的时候：
```go
r.PanicHandler = func(w http.ResponseWriter, r *http.Request, c interface{}) {
	log.Printf("Recovering from panic, Reason: %#v", c.(error))
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(c.(error).Error()))
}
```

目前开源界最为流行(star 数最多)的 web 框架 [gin](https://github.com/gin-gonic/gin) 使用的就是 httprouter 的变种。

## 原理
httprouter 和众多衍生 router 使用的数据结构被称为 radix tree，压缩字典树。读者可能没有接触过压缩字典树，但对字典树 trie tree 应该有所耳闻。下图是一个典型的字典树结构：

![trie tree](../images/ch6-02-trie.png)

字典树常用来进行字符串检索，例如用给定的字符串序列建立字典树。对于目标字符串，只要从根结点开始深度优先搜索，即可判断出该字符串是否曾经出现过，时间复杂度为 O(n)，n 可以认为是目标字符串的长度。为什么要这样做？字符串本身不像数值类型可以进行数值比较，两个字符串对比的时间复杂度取决于字符串长度。如果不用字典树来完成上述功能要复杂得多，可能要对历史字符串进行排序，时间复杂度也没有办法更低。亦可认为字典树是一种空间换时间的典型做法。

普通的字典树有一个比较明显的缺点，就是每个字母都需要建立一个孩子结点，这样会导致字典树的层树比较深，压缩字典树相对好地平衡了字典树的优点和缺点。下图是典型的压缩字典树结构：

![radix tree](../images/ch6-02-radix.png)

每个结点上不只存储一个字母了，这也是压缩字典树中“压缩”的主要含义。使用压缩字典树可以减少树的层数，同时因为每个结点上数据存储也比通常的字典树要多，所以程序的局部性较好，从而对 CPU 缓存友好。

## 压缩字典树创建过程
我们来跟踪一下 httprouter 中，一个典型的字典树的创建过程，路由设定如下：

### root 结点创建

### 子结点插入
TODO

### 子结点冲突处理

### 边分裂
