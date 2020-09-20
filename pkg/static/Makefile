GO=go
DOCKER=docker-compose
MIGRATE=migrate
CLINAME=base-project
# TODO: initialize dbs with create database on pq local

integration_test:
	$(GO) test -count=1 -tags integration ./...
unit_test:
	$(GO) test -count=1 -tags unit ./...
test:
	$(GO) test -count=1 -tags all ./...
initialize_core_db:
	${MIGRATE} -source file://db/migrations/core/ -database "postgres://postgres:postgres@localhost/core?sslmode=disable" up
initialize_db:
	${MIGRATE} -source file://db/migrations/core/ -database "postgres://postgres:postgres@localhost/core?sslmode=disable" up
local_run:
	${GO} install
	${DOCKER} up -d
	# TODO: buld docker image and use that with compose
	${MIGRATE} -source file://db/migrations/core/ -database "postgres://postgres:postgres@localhost/core?sslmode=disable" up
