.PHONY: install all

all:
	protoc --proto_path=${GOPATH}/src/ --proto_path=. \
            --go_out=${GOPATH}/src/ \
            --gogqlenum_out=${GOPATH}/src/ \
            ./*.proto

install:
	go get -u github.com/golang/protobuf/protoc-gen-go

	go install github.com/golang/protobuf/protoc-gen-go

	go install ../protobuf/protoc-gen-gogqlenum
