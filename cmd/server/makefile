BINARY_NAME=server.out
VERSION=1
DATE=$(shell date +'%Y/%m/%d %H:%M:%S')
COMMIT=$(shell git rev-parse HEAD)

build:
	go build -ldflags "-X main.buildVersion=${VERSION} -X 'main.buildDate=${DATE}' -X 'main.buildCommit=${COMMIT}'" -o ${BINARY_NAME} main.go

run:
	go build -ldflags "-X main.buildVersion=${VERSION} -X 'main.buildDate=${DATE}' -X 'main.buildCommit=${COMMIT}'" -o ${BINARY_NAME} main.go
	./${BINARY_NAME}

clean:
	go clean
	rm ${BINARY_NAME}
