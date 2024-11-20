GO = go build 

BINARY_PATH = bin
SRC_PATH = cmd

SERVER_BIN = $(BINARY_PATH)/server
CLIENT_BIN = $(BINARY_PATH)/client
STOCKFISH_BIN = $(BINARY_PATH)/stockfish

SERVER_SRC = $(SRC_PATH)/server/main.go
CLIENT_SRC = $(SRC_PATH)/client/main.go
STOCKFISH_PATH = Stockfish/src


all: $(SERVER_BIN) $(CLIENT_BIN) $(STOCKFISH_BIN)

server: $(SERVER_BIN) 

client: $(CLIENT_BIN)

run-server: $(SERVER_BIN) $(STOCKFISH_BIN)
	./$(SERVER_BIN)

run-client: $(CLIENT_BIN)
	./$(CLIENT_BIN)

$(CLIENT_BIN): $(CLIENT_SRC) $(BINARY_PATH)
	$(GO) -o $@ $<

$(SERVER_BIN): $(SERVER_SRC) $(BINARY_PATH)
	$(GO) -o $@ $<

$(STOCKFISH_BIN): $(BINARY_PATH)
	make -C $(STOCKFISH_PATH) -j profile-build
	mv $(STOCKFISH_PATH)/stockfish $(STOCKFISH_BIN)

$(BINARY_PATH):
	mkdir -p $(BINARY_PATH)

# doesn't requrie re-making stockfish
clean: $(BINARY_PATH)
	find $(BINARY_PATH) -type f ! -name "stockfish" | xargs rm -f

# removes all binaries
clean-all: $(BINARY_PATH)
	rm -f $(BINARY_PATH)/*