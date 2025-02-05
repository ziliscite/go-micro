SHELL=cmd.exe
FRONT_END_BINARY=web.exe
BROKER_BINARY=broker
AUTHENTICATION_BINARY=authentication
LOGGER_BINARY=logger
MAILER_BINARY=mailer
LISTENER_BINARY=listener

## -> those binary should match the name un the dockerfile

## up: starts all containers in the background without forcing build
.PHONY: up
up:
	@echo Starting Docker images...
	docker-compose up -d
	@echo Docker images started!

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
.PHONY: up_build
up_build: build_broker build_auth build_logger build_mailer build_listener
	@echo Stopping docker images (if running...)
	docker-compose down
	@echo Building (when required) and starting docker images...
	docker-compose up --build -d
	@echo Docker images built and started!

## down: stop docker compose
.PHONY: down
down:
	@echo Stopping docker compose...
	docker-compose down
	@echo Done!

## build_broker: builds the broker binary as a linux executable
build_broker:
	@echo Building broker binary...
	chdir ..\broker && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${BROKER_BINARY} ./cmd/api
	@echo Done!

## build_logger: builds the logger binary as a linux executable
build_logger:
	@echo Building logger binary...
	chdir ..\logger && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LOGGER_BINARY} ./cmd/api
	@echo Done!

## build_auth: builds the authentication service binary as a linux executable
build_auth:
	@echo Building authentication binary...
	chdir ..\authentication && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${AUTHENTICATION_BINARY} ./cmd/api
	@echo Done!

## build_mailer: builds the mailer binary as a linux executable
build_mailer:
	@echo Building mailer binary...
	chdir ..\mailer && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${MAILER_BINARY} ./cmd/api
	@echo Done!

## build_listener: builds the listener binary as a linux executable
build_listener:
	@echo Building listener binary...
	chdir ..\listener && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LISTENER_BINARY} .
	@echo Done!

## build_front: builds the frontend binary
build_front:
	@echo Building front end binary...
	chdir ..\frontend && set CGO_ENABLED=0&& set GOOS=windows&& go build -o ${FRONT_END_BINARY} ./cmd/web
	@echo Done!

## start: starts the front end
start: build_front
	@echo Starting front end
	chdir ..\frontend && start /B ${FRONT_END_BINARY} &

## stop: stop the front end
stop:
	@echo Stopping front end...
	@taskkill /IM "${FRONT_END_BINARY}" /F
	@echo "Stopped front end!"

