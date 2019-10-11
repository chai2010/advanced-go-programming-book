generate:
	protoc --go_out=. hello.proto

generate-netrpc:
	protoc --go-netrpc_out=plugins=netrpc:. hello.proto

clean:
	-rm *.pb.go

