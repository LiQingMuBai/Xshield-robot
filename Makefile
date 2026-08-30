BINARY_NAME=ushield_bot
BINARY_DIR=bin
MAIN_PKG=./cmd

.PHONY: all build build-linux build-darwin build-windows clean run test vet tidy fmt help

all: build

build:
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PKG)

build-linux:
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)_linux_amd64 $(MAIN_PKG)

build-darwin:
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_DIR)/$(BINARY_NAME)_darwin_arm64 $(MAIN_PKG)
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DIR)/$(BINARY_NAME)_darwin_amd64 $(MAIN_PKG)

build-windows:
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)_windows_amd64.exe $(MAIN_PKG)

release: build-linux build-darwin build-windows

clean:
	rm -rf $(BINARY_DIR)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

run: build
	./$(BINARY_DIR)/$(BINARY_NAME)

test:
	go test ./... -v -count=1

vet:
	go vet ./...

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

help:
	@echo "Available targets:"
	@echo "  make build          - Build binary for current platform into bin/"
	@echo "  make build-linux    - Cross-compile for Linux amd64"
	@echo "  make build-darwin   - Cross-compile for macOS (arm64 + amd64)"
	@echo "  make build-windows  - Cross-compile for Windows amd64"
	@echo "  make release        - Build binaries for all platforms"
	@echo "  make clean          - Remove built binaries"
	@echo "  make run            - Build and run the bot"
	@echo "  make test           - Run all tests"
	@echo "  make vet            - Run go vet"
	@echo "  make tidy           - Run go mod tidy"
	@echo "  make fmt            - Format Go source files"
