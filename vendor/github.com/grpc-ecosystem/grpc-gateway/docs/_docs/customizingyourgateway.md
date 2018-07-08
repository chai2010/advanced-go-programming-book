---
title: Customizing your gateway
category: documentation
order: 101
---

# Customizing your gateway

## Message serialization

You might want to serialize request/response messages in MessagePack instead of JSON, for example.

1. Write a custom implementation of [`Marshaler`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#Marshaler)
2. Register your marshaler with [`WithMarshalerOption`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#WithMarshalerOption)
   e.g.
   ```go
   var m your.MsgPackMarshaler
   mux := runtime.NewServeMux(runtime.WithMarshalerOption("application/x-msgpack", m))
   ```

You can see [the default implementation for JSON](https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/marshal_jsonpb.go) for reference.

## Mapping from HTTP request headers to gRPC client metadata
You might not like [the default mapping rule](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#DefaultHeaderMatcher) and might want to pass through all the HTTP headers, for example.

1. Write a [`HeaderMatcherFunc`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#HeaderMatcherFunc).
2. Register the function with [`WithIncomingHeaderMatcher`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#WithIncomingHeaderMatcher)

   e.g.
   ```go
   func yourMatcher(headerName string) (mdName string, ok bool) {
   	...
   }
   ...
   mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(yourMatcher))

   ```

## Mapping from gRPC server metadata to HTTP response headers
ditto. Use [`WithOutgoingHeaderMatcher`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#WithOutgoingHeaderMatcher)

## Mutate response messages or set response headers
You might want to return a subset of response fields as HTTP response headers; 
You might want to simply set an application-specific token in a header.
Or you might want to mutate the response messages to be returned.

1. Write a filter function.
   ```go
   func myFilter(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
   	w.Header().Set("X-My-Tracking-Token", resp.Token)
   	resp.Token = ""
   	return nil
   }
   ```
2. Register the filter with [`WithForwardResponseOption`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#WithForwardResponseOption)
   
   e.g.
   ```go
   mux := runtime.NewServeMux(runtime.WithForwardResponseOption(myFilter))
   ```

## Error handler
http://mycodesmells.com/post/grpc-gateway-error-handler

## Replace a response forwarder per method
You might want to keep the behavior of the current marshaler but change only a message forwarding of a certain API method.

1. write a custom forwarder which is compatible to [`ForwardResponseMessage`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#ForwardResponseMessage) or [`ForwardResponseStream`](http://godoc.org/github.com/grpc-ecosystem/grpc-gateway/runtime#ForwardResponseStream).
2. replace the default forwarder of the method with your one.

   e.g. add `forwarder_overwrite.go` into the go package of the generated code,
   ```go
   package generated
   
   import (
   	"net/http"

   	"github.com/grpc-ecosystem/grpc-gateway/runtime"
   	"github.com/golang/protobuf/proto"
   	"golang.org/x/net/context"
   )

   func forwardCheckoutResp(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, resp proto.Message, opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
   	if someCondition(resp) {
   		http.Error(w, "not enough credit", http. StatusPaymentRequired)
   		return
   	}
   	runtime.ForwardResponseMessage(ctx, mux, marshaler, w, req, resp, opts...)
   }
   
   func init() {
   	forward_MyService_Checkout_0 = forwardCheckoutResp
   }
   ```
