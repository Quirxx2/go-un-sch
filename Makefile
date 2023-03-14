MAKEFLAGS += --silent

.PHONY: clean
clean: 
	rm -rf ./tmp

.PHONY: build.docker
build.docker:
	docker buildx build -t=certs --target=release .

.PHONY: build.proto
build.proto: build.proto.requires
	protoc -I=./api -I=./api/include \
	--go_out=./api --go_opt=paths=source_relative \
	--go-grpc_out=./api --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=./api --grpc-gateway_opt=paths=source_relative \
	--grpc-gateway_opt=grpc_api_configuration=./api/certs.yaml \
	certs.proto 

.PHONY: build.proto.requires
build.proto.requires:
	if ! type protoc > /dev/null; then \
		echo "protoc required: https://grpc.io/docs/languages/go/quickstart/#prerequisites"; \
		exit 1; \
	fi;
	if ! type protoc-gen-grpc-gateway > /dev/null; then \
		echo "grpc-gateway required: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"; \
		exit 1; \
	fi;

.PHONY: build.proto.get_imports
build.proto.get_imports:
	wget -N -P ./api/include/google/api/ \
		https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/httpbody.proto \
		https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto \
		https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto

.PHONY: build.mocks
build.mocks: build.mocks.requires build.mock.Registry build.mock.Storage build.mock.Templater

.PHONY: build.mocks.requires
build.mocks.requires:
	if ! type mockery > /dev/null; then \
  		echo "mockery required: go install github.com/vektra/mockery/v2@latest"; \
		exit 1;\
	fi;

.PHONY: build.mock.Registry
build.mock.Registry: build.mocks.requires
	mockery --name=Registry --inpackage --testonly --case underscore --with-expecter;

.PHONY: build.mock.Storage
build.mock.Storage: build.mocks.requires
	mockery --name=Storage --inpackage --testonly --case underscore --with-expecter;

.PHONY: build.mock.Templater
build.mock.Templater: build.mocks.requires
	mockery --name=Templater --inpackage --testonly --case underscore --with-expecter;

TEST_COMPOSE=export HOST_UID=$$(id -u):$$(id -g); docker compose -f docker-compose.yml -f docker-compose.test.yml
.PHONY: up
up: 
	$(TEST_COMPOSE) up -d -V

.PHONY: down
down:
	docker compose down 

.PHONY: restart
restart: down up

.PHONY: db.up
db.up:
	$(TEST_COMPOSE) up -d -V db

.PHONY: db.down
db.down:
	docker compose down 

.PHONY: db.wait
db.wait: db.up
	echo "waiting for postgres"; 
	docker compose exec db sh -c 'until pg_isready -q; do \
								  	{ printf .; sleep 0.1; }; \
								  done;'
	echo "\\npostgres is ready";
	sleep 1;

.PHONY: gotenberg.up
gotenberg.up:
	$(TEST_COMPOSE) up -d -V gotenberg

.PHONY: gotenberg.down
gotenberg.down:
	docker compose down 

INTEGRATION_PATH=./test/integration/
.PHONY: test.it.all
test.it.all: test.it.db.all test.it.gotenberg.all

.PHONY: test.it.db
test.it.db: db.down db.wait
	echo "running integration test: $(t)"
	go test -v -count=1 -tags integration $(INTEGRATION_PATH)$(t) 

.PHONY: test.it.db.all
test.it.db.all:
	for file in `find $(INTEGRATION_PATH) -name 'db*_test.go' -type f`; do \
		make test.it.db t=`basename $$file`; \
	done;
	make db.down

.PHONY: test.it.gotenberg
test.it.gotenberg: gotenberg.down gotenberg.up
	echo "running integration test: $(t)"
	go test -v -count=1 -tags integration $(INTEGRATION_PATH)$(t) 

.PHONY: test.it.gotenberg.all
test.it.gotenberg.all:
	for file in `find $(INTEGRATION_PATH) -name 'gotenberg*_test.go' -type f`; do \
		make test.it.gotenberg t=`basename $$file`; \
	done;
	make gotenberg.down
	
E2E_PATH=./test/e2e/
E2E_OUT_PATH=./tmp/e2e/
.PHONY: test.e2e
test.e2e: test.e2e.requires down up
	echo "running integration test: $(t)"
	sleep 1
	go test -v -count=1 -tags e2e $(E2E_PATH)$(t)

.PHONY: test.e2e.requires
test.e2e.requires:
	mkdir -p $(E2E_OUT_PATH)

.PHONY: test.e2e.all
test.e2e.all:
	for file in `find $(E2E_PATH) -name 'e2e*_test.go' -type f`; do \
		make test.e2e t=`basename $$file`; \
	done;
	make down