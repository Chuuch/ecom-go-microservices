.PHONY: local

local:
	echo "Starting local docker compose"
	docker compose -f docker-compose.yaml up -d --build

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)