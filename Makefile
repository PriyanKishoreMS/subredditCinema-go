all: build

build:
	@echo building
	@go build -o main cmd/*

run:
	@go run cmd/*

watch:
	@air

mysql:
	@docker compose up mysql

psql:
	@docker compose up psql

down:
	@docker compose down	

dsn := $(shell cat dsnpsql.txt)

goose_create:
	@goose -s -dir='./migrations' postgres "${dsn}" create "${fn}" sql

goose_one:
	@goose -dir='./migrations' postgres "${dsn}" up-by-one

goose_down:
	@goose -dir='./migrations' postgres "${dsn}" down

goose_up:
	@goose -dir='./migrations' postgres "${dsn}" up

