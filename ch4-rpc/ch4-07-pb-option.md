# 4.7. Protobuf扩展语法和插件

TODO

<!--

基于pb扩展，打造一个自定义的rest生成

支持 url 和 url.Values

通过 grpc-gateway/runtime.PopulateFieldFromPath 和 PopulateQueryParameters 天才 protoMsg 成员

路由通过 httprouter 处理

- https://github.com/julienschmidt/httprouter
- https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/query.go#L20

先生成 net/rpc 接口，然后同时增加 Rest 接口

扩展的元信息需要一个独立的文件，因为在插件中需要访问。

可以新开一个github项目，便于引用

-->
