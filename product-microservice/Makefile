.PHONY: local proto

local:
	echo "Starting local docker compose"
	docker compose -f docker-compose.yaml up -d --build

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

proto:
	protoc --go_out=./proto/product --go-grpc_out=./proto/product --proto_path=./proto proto/product/product.proto