BINARY = main
SRC = main.go

.PHONY: build
build: $(BINARY)

$(BINARY): $(SRC)
	echo "Building $(BINARY) ..."
	go build -ldflags "-X main.Version=$(shell date +'%Y%m%d.%H%M%S')" -o ./$(BINARY)


clear:
	rm -rf bin/*
	rm -rf out/*