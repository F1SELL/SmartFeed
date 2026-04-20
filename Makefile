.PHONY: compose-up compose-down run-core run-feed-worker run-ai-worker migrate-up migrate-down kafka-init test race lint build format format-check ci load-test load-test-docker swagger

# Поднять инфраструктуру (базы данных, брокер)
compose-up:
	docker compose up -d

# Опустить инфраструктуру
compose-down:
	docker compose down -v

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

kafka-init:
	docker exec smartfeed-kafka /opt/kafka/bin/kafka-topics.sh --create --if-not-exists --topic post_created --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1

run-core:
	go run ./cmd/core

run-feed-worker:
	go run ./cmd/feed-worker

run-ai-worker:
	go run ./cmd/ai-worker

test:
	go test ./...

race:
	go test -race ./...

lint:
	go vet ./...

build:
	go build ./...

format:
	gofmt -w $$(find . -name '*.go' -type f)

format-check:
	@out="$$(gofmt -l $$(find . -name '*.go' -type f))"; \
	if [ -n "$$out" ]; then \
		echo "Unformatted files:"; \
		echo "$$out"; \
		exit 1; \
	fi

ci: format-check lint test race build

swagger:
	swag init -g cmd/core/main.go --parseDependency --parseInternal

load-test: kafka-init
	k6 run -e BASE_URL=http://localhost:8080 ./load/k6/feed.js

load-test-docker: kafka-init
	docker run --rm -i -v "$(PWD)/load/k6:/scripts" grafana/k6:0.49.0 run -e BASE_URL=http://host.docker.internal:8080 /scripts/feed.js
