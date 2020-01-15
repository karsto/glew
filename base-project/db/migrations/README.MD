# Simple Migrations Overview

[Tool](https://github.com/golang-migrate/migrate)

DB Setup:
```
// create db
// create schema
// create user + permissions
CREATE EXTENSION IF NOT EXISTS postgis;
```

Migrate CLI Setup for postgres:
```
go get -u -d github.com/golang-migrate/migrate/cli github.com/lib/pq
go build -tags 'postgres' -o /usr/local/bin/migrate github.com/golang-migrate/migrate/cli
```

Migration command- additional [options](https://github.com/golang-migrate/migrate/tree/master/database/postgres)
```
migrate -source file://{path} -database "postgres://{username}:{password}@{ip}/{dbName}?sslmode=disable" up {version}
migrate -source file://db/migrations/core/ -database "postgres://postgres:postgres@localhost/core?sslmode=disable" up
```
