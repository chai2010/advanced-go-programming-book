# 4.7. Protobuf扩展语法和插件

在本章第二节我们已经展示过如何定制一个Protobuf代码生成插件。本节我们将继续深入挖掘Protobuf的高级特性，通过Protobuf的扩展特性增加自定义的元数据信息。通过针对每个方法增加Rest接口元信息，实现一个基于Protobuf的迷你Rest框架。

## Protobuf扩展语法

目前Protobuf相关的很多开源项目都使用到了Protobuf的扩展语法。在前一节中提到的验证器就是通过给结构体成员增加扩展元信息实现验证。在grpc-gateway项目中，则是通过为服务的每个方法增加Http相关的映射规则实现对Rest接口的支持。这里我们将查看下Protobuf全部的扩展语法。

扩展语法也被用来实现Protobuf内置的某些特性，比如针对不同语言的扩展选项和proto2中message成员的默认值特性：其中文件扩展选项go_package为Go语言定义了当前包的路径和包的名称。message成员的default扩展为每个成员定义了默认值。go_package和proto2的default特性底层都是通过扩展语法实现。

Protobuf扩展语法有五种类型，分别是针对文件的扩展信息、针对message的扩展信息、正对message成员的扩展信息、针对service的扩展信息和针对service方法的扩展信息。在使用扩展前首先需要通过extend关键字定义扩展的类型和可以用于扩展的成员。扩展成员也可以基础类型，也可以是一个结构体类型。

为了简单，我们假设采用标准库中的StringValue作为每个扩展成员的类型：

```protobuf
import "google/protobuf/wrappers.proto";

message StringValue {
	// The string value.
	string value = 1;
}
```

我们先看看如何定义文件的扩展类型：

```protobuf
import "google/protobuf/descriptor.proto";

extend google.protobuf.FileOptions {
	optional google.protobuf.StringValue file_option = 50000;
}

option (file_option) = {
	value: "this is a file option"
};
```

然后是message和message成员的扩展方式：

```protobuf
import "google/protobuf/descriptor.proto";

extend google.protobuf.MessageOptions {
	optional google.protobuf.StringValue message_option = 50000;
}
extend google.protobuf.FieldOptions {
	optional google.protobuf.StringValue filed_option = 50000;
}

message Message {
	option (message_option) = {
		value: "message option"
	};

	string name = 1 [
		(filed_option) = {
			value: ""
		}
	];
}
```

最后是service和service方法的扩展：

```protobuf
import "google/protobuf/descriptor.proto";

extend google.protobuf.ServiceOptions {
	optional String service_option = 50000;
}
extend google.protobuf.MethodOptions {
	optional String method_option = 50000;
}

service HelloService {
	option (service_option) = {
		value: "message option"
	};

	rpc Hello(String) returns(String) {
		option (method_option) = {
			value: ""
		};
	}
}
```

如果是通用的扩展类型，我们可以将扩展相关的内容放到一个独立的proto文件中。以后知道导入定义了扩展的proto文件，就可以直接使用定义的扩展定义元数据了。

## 插件中读取扩展信息

TODO

## Rest模板框架

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
