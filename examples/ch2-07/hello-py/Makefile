default:
	go build -buildmode=c-shared -o gopkg.so main.go
	python3 -c 'import gopkg; print(gopkg.system("time"))'
	python3 -c 'import gopkg; print(gopkg.sum(2, 3))'

clean:
	-rm *.so

# PKG_CONFIG=python3-config go build -buildmode=c-shared -o say-hello.so main.go
# python3-config --ldflags
# python3-config --include

