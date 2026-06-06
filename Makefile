BINARY    := pin
BIN_DIR   := bin
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: all install-bin clean

all: $(BIN_DIR)/$(BINARY)

$(BIN_DIR)/$(BINARY): $(BIN_DIR) cmd/pin/main.go
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/pin

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

install-bin: $(BIN_DIR)/$(BINARY)
	mkdir -p $(INSTALL_DIR)
	install -m 755 $(BIN_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY)"

clean:
	rm -rf $(BIN_DIR)
