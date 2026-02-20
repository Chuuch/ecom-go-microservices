.PHONY: local proto

local:
	echo "Starting local docker compose"
	mkdir -p data/kafka1/data data/kafka2/data data/kafka3/data
	sudo chown -R $(shell id -u):$(shell id -g) data/
	docker compose -f docker-compose.yaml up -d --build

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

proto:
	protoc --go_out=./proto/product --go-grpc_out=./proto/product --proto_path=./proto proto/product/product.proto

create_topic:
#	docker exec -it kafka1 kafka-topics --bootstrap-server kafka1:19091 --create --topic create_product --partitions 3 --replication-factor 2
#	docker exec -it kafka1 kafka-topics --bootstrap-server kafka1:19091 --create --topic update_product --partitions 3 --replication-factor 2
	docker exec -it kafka1 kafka-topics --bootstrap-server kafka1:19091 --create --topic dead_letter_queue --partitions 3 --replication-factor 2

down:
	rm -rf data/kafka1/data data/kafka2/data data/kafka3/data
	docker compose -f docker-compose.yaml down -v
