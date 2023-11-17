posada:
	docker run --name posada --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it posada createdb --username=root --owner=root posada

migrateup:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/posada?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/posada?sslmode=disable" -verbose down

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/newbri/posadamissportia/db Store

.PHONY: todo createdb migrateup migratedown mock