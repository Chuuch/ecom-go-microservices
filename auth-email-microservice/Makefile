run:
	config=local go run ./cmd/auth/main.go

run-docker:
	config=docker go run ./cmd/auth/main.go

build:
	go build -o ./bin/ecom-microservices ./cmd/auth/main.go

proto:
	protoc --go_out=./proto --go-grpc_out=./proto --proto_path=./proto ./proto/*.proto

docker:
	echo "Starting docker containers"
	docker run -d --name jaeger \
		-e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
		-p 5775:5775/udp \
		-p 6831:6831/udp \
		-p 6832:6832/udp \
		-p 5778:5778 \
		-p 16686:16686 \
		-p 14268:14268 \
		-p 14250:14250 \
		-p 9411:9411 \
		jaegertracing/all-in-one:1.72.0
	docker-compose -f docker-compose.yaml up --build

docker-compose:
	docker-compose -f docker-compose.yaml up --build

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

migrate-up:
	migrate -database postgres://postgres:postgres@localhost:5432/ecom_auth_db?sslmode=disable -path ./migrations up 1

migrate-down:
	migrate -database postgres://postgres:postgres@localhost:5432/ecom_auth_db?sslmode=disable -path ./migrations down 1

migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(name)

version:
	migrate -database postgres://postgres:postgres@localhost:5432/ecom_auth_db?sslmode=disable -path ./migrations version

force:
	migrate -database postgres://postgres:postgres@localhost:5432/ecom_auth_db?sslmode=disable -path ./migrations force 1

test:
	go test -cover ./...

.PHONY: proto docker 