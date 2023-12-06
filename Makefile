BINARY = bin/a.out
SRC = src/*.go

.PHONY: build
build: $(BINARY)
$(BINARY): $(SRC)
	@echo "Building $(BINARY) ..."
	cd src && go build -o ../$(BINARY)


clear:
	rm -rf bin/*
	rm -rf out/*