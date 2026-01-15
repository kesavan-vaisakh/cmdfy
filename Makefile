# Binary name
BINARY_NAME=cmdfy
ENTRY_POINT=app/main.go

# Build directory
BUILD_DIR=bin

.PHONY: all clean build build-mac build-linux build-windows

all: clean build-all

build:
	@echo "Building for current OS..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(ENTRY_POINT)

build-all: build-mac build-linux build-windows

build-mac:
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-mac-amd64 $(ENTRY_POINT)
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-mac-arm64 $(ENTRY_POINT)

build-linux:
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(ENTRY_POINT)
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(ENTRY_POINT)

build-windows:
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(ENTRY_POINT)

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
