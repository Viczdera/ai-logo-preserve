
postgres:
	docker run --name logo_preserve_container -p 8080:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it logo_preserve_container createdb --username=root --owner=root logo_preserve_db

dropdb:
	docker exec -it logo_preserve_container dropdb logo_preserve_db

migrate-up:
	migrate -path internal/db/migrations -database postgres://root:secret@localhost:8080/logo_preserve_db?sslmode=disable up

migrate-down:
	migrate -path internal/db/migrations -database postgres://root:secret@localhost:8080/logo_preserve_db?sslmode=disable down

rabbitmq:
	docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:alpine

produce-msg:
	go run internal/queue/send/main.go

consume-msg:
	go run internal/queue/receive/main.go

server:
	go run main.go