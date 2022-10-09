BINARY=$(shell basename $(CURDIR))

.PHONY: all
all: build

.PHONY: build
build:
	go build -o ${BINARY}

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -f ./${BINARY}