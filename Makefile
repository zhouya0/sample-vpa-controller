IMAGE_NAME=daocloud.io/dcelib/vpa-controller
VERSION=v1.0.0

all: publish

build-linux:
	GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o=vpa main.go

publish: build-linux
	docker build --no-cache -t ${IMAGE_NAME}:${VERSION} .
	docker push ${IMAGE_NAME}:${VERSION}

