BINARY_DIR = ./bin
SERVER_BINARY = $(BINARY_DIR)/server
AGENT_BINARY = $(BINARY_DIR)/agent
GO = go

.PHONY: all clean client agent

all: client agent

$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

client: $(BINARY_DIR)
	$(GO) build -o $(SERVER_BINARY) ./cmd/server

agent: $(BINARY_DIR)
	$(GO) build -o $(AGENT_BINARY) ./cmd/agent

clean:
	rm -rf $(BINARY_DIR)

rebuild: clean all

run_test: test_server test_client

test_server: $(SERVER_BINARY)
	./metricstest -test.v -test.run=TestIteration2 -binary-path=$(SERVER_BINARY)

test_client: $(AGENT_BINARY)
	./metricstest -test.v -test.run=TestIteration2 -binary-path=$(AGENT_BINARY)