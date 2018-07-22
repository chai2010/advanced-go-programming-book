
run:
	@go build -o a.out && ./a.out
	-@rm ./a.out

gen:
	protoc \
		-I=. -I=../../../github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:. \
		helloworld.proto
	protoc \
		-I=. -I=../../../github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=. \
		helloworld.proto
	protoc \
		-I=. -I=../../../github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=. \
		helloworld.proto

clean:
	-rm *.pb.go
