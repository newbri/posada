todo:
	docker run --name todo --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it todo createdb --username=root --owner=root todo

sqlc:
	sqlc generate

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/newbri/todo/db/sqlc Store

.PHONY: todo createdb sqlc mock