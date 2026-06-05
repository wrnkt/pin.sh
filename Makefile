BINARY  := pin
BIN_DIR := bin

.PHONY: all clean

all: $(BIN_DIR)/$(BINARY)

$(BIN_DIR)/$(BINARY): $(BIN_DIR) cmd/pin/main.go
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/pin

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

clean:
	rm -rf $(BIN_DIR)
