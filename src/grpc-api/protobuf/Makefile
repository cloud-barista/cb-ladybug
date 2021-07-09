regenerate:
	# gogo proto compile
	protoc \
			./gogoproto/gogo.proto \
			-I . \
			-I $(GOPATH)/src/github.com/gogo/protobuf/protobuf \
			-I $(GOPATH)/src \
			--gofast_out=plugins=grpc,paths=source_relative,\
	Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:\
	.

	# cbladybug proto compile
	protoc \
			./cbladybug/cbladybug.proto \
			-I . \
			-I $(GOPATH)/src/github.com/gogo/protobuf/protobuf \
			-I $(GOPATH)/src/github.com/cloud-barista/cb-ladybug/src/grpc-api/protobuf \
			--gofast_out=plugins=grpc:\
	.	
