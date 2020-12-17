GOPATH ?= ${HOME}/go
MODCACHE ?= ${GOPATH}/pkg/mod

DOCKER_UID = $(shell id -u)
DOCKER_GID = $(shell id -g)

CUR_DIR := $(shell pwd)

BUILD_IMAGE := unanet-docker.jfrog.io/golang

docker-exec = docker run --rm \
	-e DOCKER_UID=${DOCKER_UID} \
	-e DOCKER_GID=${DOCKER_GID} \
	-v ${CUR_DIR}:/src \
	-v ${MODCACHE}:/go/pkg/mod \
	-v ${HOME}/.ssh/id_rsa:/home/unanet/.ssh/id_rsa \
	-w /src \
	${BUILD_IMAGE}

.PHONY: build dist test swagger

test:
	docker pull ${BUILD_IMAGE}
	$(docker-exec) go build ./...
	$(docker-exec) go vet ./...
	$(docker-exec) go test -tags !local ./...

