AGENT_PATH = cmd/agent/main.go
SERVER_PATH = cmd/server/main.go

BUILD_DIR = build
AGENT_BUILD_DIR = $(BUILD_DIR)/agent
SERVER_BUILD_DIR = $(BUILD_DIR)/server

.PHONY: all
all: agent server_amd64 server_arm64 server_darwin compress

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)
$(AGENT_BUILD_DIR):
	mkdir -p $(AGENT_BUILD_DIR)
$(SERVER_BUILD_DIR):
	mkdir -p $(SERVER_BUILD_DIR)

.PHONY: agent
agent: $(AGENT_BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(AGENT_BUILD_DIR)/convoC2_agent.exe $(AGENT_PATH)
	@echo "Built agent for Windows (amd64)"

.PHONY: server_amd64
server_amd64: $(SERVER_BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_amd64 $(SERVER_PATH)
	@echo "Built server for Linux (amd64)"

.PHONY: server_arm64
server_arm64: $(SERVER_BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_arm64 $(SERVER_PATH)
	@echo "Built server for Linux (arm64)"

.PHONY: server_darwin
server_darwin: $(SERVER_BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_darwin $(SERVER_PATH)
	@echo "Built server for MacOS (arm64)"

.PHONY: compress
compress: agent server_amd64 server_arm64 server_darwin
	@echo "Compressing build outputs..."
	tar -czf $(AGENT_BUILD_DIR)/convoC2_agent.tar.gz -C $(AGENT_BUILD_DIR) convoC2_agent.exe
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_amd64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_amd64
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_arm64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_arm64
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_darwin.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_darwin
	@echo "Compression complete"

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build directories"
