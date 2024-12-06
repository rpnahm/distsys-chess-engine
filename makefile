GO = go build 

BINARY_PATH = bin
SRC_PATH = cmd

SERVER_BIN = $(BINARY_PATH)/server
CLIENT_BIN = $(BINARY_PATH)/client
STOCKFISH_BIN = $(BINARY_PATH)/stockfish
TEST_BIN = $(BINARY_PATH)/test

SERVER_SRC = $(SRC_PATH)/server/main.go
CLIENT_SRC = $(SRC_PATH)/client/main.go
TEST_SRC = $(SRC_PATH)/test/main.go
STOCKFISH_PATH = Stockfish/src

UTILS = pkg


all: $(SERVER_BIN) $(CLIENT_BIN) $(STOCKFISH_BIN)

server: $(SERVER_BIN) 

client: $(CLIENT_BIN)

test: $(TEST_BIN)

run-server: $(SERVER_BIN)
	./$(SERVER_BIN) test-rnahm-00

run-client: $(CLIENT_BIN)
	./$(CLIENT_BIN) test-rnahm

run-test: $(TEST_BIN)
	./$(TEST_BIN) rnahm 2 3 10 1

$(CLIENT_BIN): $(CLIENT_SRC) $(UTILS)/client/* $(UTILS)/common/* $(BINARY_PATH)
	$(GO) -o $@ $<

$(SERVER_BIN): $(SERVER_SRC) $(UTILS)/server/* $(UTILS)/common/* $(BINARY_PATH)
	$(GO) -o $@ $<

$(TEST_BIN): $(TEST_SRC) $(UTILS)/client/* $(UTILS)/common/* $(BINARY_PATH)
	$(GO) -o $@ $<
	
$(STOCKFISH_BIN): $(BINARY_PATH)
	make -C $(STOCKFISH_PATH) -j profile-build
	cp $(STOCKFISH_PATH)/stockfish $(STOCKFISH_BIN)

$(BINARY_PATH):
	mkdir -p $(BINARY_PATH)

# doesn't requrie re-making stockfish
clean: $(BINARY_PATH)
	find $(BINARY_PATH) -type f ! -name "stockfish" | xargs rm -f

# removes all binaries
clean-all: $(BINARY_PATH)
	rm -f $(BINARY_PATH)/*