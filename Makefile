.PHONY: build run clean docker-build docker-run test vet fmt

IMAGE        ?= priority-stack
PORT_PUBLISH ?= 8095
DATA_FILE    ?= ./projects.json
PORT         ?= 8080

build:
	go build -trimpath -ldflags="-s -w" -o priority .

run: build
	DATA_FILE=$(DATA_FILE) PORT=$(PORT) ./priority

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -f priority priority.exe

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p $(PORT_PUBLISH):8080 -v /volume1/docker/priority/data:/app/data $(IMAGE)
