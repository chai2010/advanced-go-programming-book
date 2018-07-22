
default:
	openssl genrsa -out server.key 2048
	openssl req -new -x509 -days 3650 \
		-subj "/C=GB/L=China/O=gobook/CN=server.grpc.io" \
		-key server.key -out server.crt

clean:
	-rm *.key *.crt *.csr *.srl

