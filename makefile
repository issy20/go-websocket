APPNAME="go-simple-chat"
OUTDIR="./dist"
DSN="mysql://$(MYSQL_DSN)"
MIGRATION_DIR="file://db/migrations"

depend:
	@go mod tidy
	@go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

run: depend
	@go get github.com/pilu/fresh
	fresh

# データベースのバージョンを１つ進めます。
up: depend
	migrate -database=$(DSN) -source=$(MIGRATION_DIR) up

# データベースのバージョンを１つ戻します。
down: depend
	migrate -database=$(DSN) -source=$(MIGRATION_DIR) down

drop:
	migrate -database=$(DSN) -source=$(MIGRATION_DIR) drop
