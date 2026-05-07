BIN := bin/lt-clean

.PHONY: build build-small install test clean fmt vet tidy

build:
	go build -o $(BIN) .

build-small:
	go build -ldflags="-s -w" -trimpath -o $(BIN) .
	@command -v upx >/dev/null && upx --best --lzma $(BIN) || echo "(upx not installed, skipping compression)"

install: build-small
	cp $(BIN) /usr/local/bin/lt-clean

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ dist/
