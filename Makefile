APP_VERSION=$(shell cicd/version.sh)
BINARY_NAME=ydbops
TODAY=$(shell date --iso=minutes)

build:
	go get -u
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build  -ldflags='-X main.buildInfo=${TODAY}' -o bin/${BINARY_NAME} main.go 
	GOOS=darwin GOARCH=amd64 go build  -ldflags='-X main.buildInfo=${TODAY}' -o bin/${BINARY_NAME}_darwin_amd64 main.go
	GOOS=darwin GOARCH=arm64 go build  -ldflags='-X main.buildInfo=${TODAY}' -o bin/${BINARY_NAME}_darwin_arm64 main.go
	strip bin/${BINARY_NAME}
clear:
	rm -rf bin/${BINARY_NAME}

dep:
	go mod download
docker:
	docker build --force-rm -t $(BINARY_NAME) .
build-in-docker: docker
	docker rm -f $(BINARY_NAME) || true
	docker create --name $(BINARY_NAME) $(BINARY_NAME)
	docker cp '$(BINARY_NAME):/app/' $(BUILD_DIR)
	docker rm -f $(BINARY_NAME)

