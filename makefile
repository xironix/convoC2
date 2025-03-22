AGENT_PATH = cmd/agent/main.go
SERVER_PATH = cmd/server/main.go

BUILD_DIR = build
AGENT_BUILD_DIR = $(BUILD_DIR)/agent
SERVER_BUILD_DIR = $(BUILD_DIR)/server

.PHONY: all
all: agents servers compress

.PHONY: agents
agents: agent_windows_amd64 agent_windows_arm64 agent_linux_amd64 agent_linux_arm64 agent_darwin_amd64 agent_darwin_arm64

.PHONY: servers
servers: server_windows_amd64 server_windows_arm64 server_linux_amd64 server_linux_arm64 server_darwin_amd64 server_darwin_arm64

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)
$(AGENT_BUILD_DIR):
	mkdir -p $(AGENT_BUILD_DIR)
$(SERVER_BUILD_DIR):
	mkdir -p $(SERVER_BUILD_DIR)

# Windows Agents
.PHONY: agent_windows_amd64
agent_windows_amd64: $(AGENT_BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(AGENT_BUILD_DIR)/convoC2_agent_windows_amd64.exe $(AGENT_PATH)
	@echo "Built agent for Windows (amd64)"

.PHONY: agent_windows_arm64
agent_windows_arm64: $(AGENT_BUILD_DIR)
	GOOS=windows GOARCH=arm64 go build -o $(AGENT_BUILD_DIR)/convoC2_agent_windows_arm64.exe $(AGENT_PATH)
	@echo "Built agent for Windows (arm64)"

# Linux Agents
.PHONY: agent_linux_amd64
agent_linux_amd64: $(AGENT_BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(AGENT_BUILD_DIR)/convoC2_agent_linux_amd64 $(AGENT_PATH)
	@echo "Built agent for Linux (amd64)"

.PHONY: agent_linux_arm64
agent_linux_arm64: $(AGENT_BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(AGENT_BUILD_DIR)/convoC2_agent_linux_arm64 $(AGENT_PATH)
	@echo "Built agent for Linux (arm64)"

# MacOS Agents
.PHONY: agent_darwin_amd64
agent_darwin_amd64: $(AGENT_BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(AGENT_BUILD_DIR)/convoC2_agent_darwin_amd64 $(AGENT_PATH)
	@echo "Built agent for MacOS (amd64)"

.PHONY: agent_darwin_arm64
agent_darwin_arm64: $(AGENT_BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(AGENT_BUILD_DIR)/convoC2_agent_darwin_arm64 $(AGENT_PATH)
	@echo "Built agent for MacOS (arm64)"

# Windows Servers
.PHONY: server_windows_amd64
server_windows_amd64: $(SERVER_BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_windows_amd64.exe $(SERVER_PATH)
	@echo "Built server for Windows (amd64)"

.PHONY: server_windows_arm64
server_windows_arm64: $(SERVER_BUILD_DIR)
	GOOS=windows GOARCH=arm64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_windows_arm64.exe $(SERVER_PATH)
	@echo "Built server for Windows (arm64)"

# Linux Servers
.PHONY: server_linux_amd64
server_linux_amd64: $(SERVER_BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_linux_amd64 $(SERVER_PATH)
	@echo "Built server for Linux (amd64)"

.PHONY: server_linux_arm64
server_linux_arm64: $(SERVER_BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_linux_arm64 $(SERVER_PATH)
	@echo "Built server for Linux (arm64)"

# MacOS Server
.PHONY: server_darwin_amd64
server_darwin_amd64: $(SERVER_BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_darwin_amd64 $(SERVER_PATH)
	@echo "Built server for MacOS (amd64)"

.PHONY: server_darwin_arm64
server_darwin_arm64: $(SERVER_BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(SERVER_BUILD_DIR)/convoC2_server_darwin_arm64 $(SERVER_PATH)
	@echo "Built server for MacOS (arm64)"

.PHONY: compress
compress: agents servers
	@echo "Compressing build outputs..."
	# Compress Windows agents
	tar -czf $(AGENT_BUILD_DIR)/convoC2_agent_windows_amd64.tar.gz -C $(AGENT_BUILD_DIR) convoC2_agent_windows_amd64.exe
	tar -czf $(AGENT_BUILD_DIR)/convoC2_agent_windows_arm64.tar.gz -C $(AGENT_BUILD_DIR) convoC2_agent_windows_arm64.exe
	# Compress Linux agents
	tar -czf $(AGENT_BUILD_DIR)/convoC2_agent_linux_amd64.tar.gz -C $(AGENT_BUILD_DIR) convoC2_agent_linux_amd64
	tar -czf $(AGENT_BUILD_DIR)/convoC2_agent_linux_arm64.tar.gz -C $(AGENT_BUILD_DIR) convoC2_agent_linux_arm64
	# Compress MacOS agents
	tar -czf $(AGENT_BUILD_DIR)/convoC2_agent_darwin_amd64.tar.gz -C $(AGENT_BUILD_DIR) convoC2_agent_darwin_amd64
	tar -czf $(AGENT_BUILD_DIR)/convoC2_agent_darwin_arm64.tar.gz -C $(AGENT_BUILD_DIR) convoC2_agent_darwin_arm64
	# Compress Windows servers
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_windows_amd64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_windows_amd64.exe
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_windows_arm64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_windows_arm64.exe
	# Compress Linux servers
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_linux_amd64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_linux_amd64
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_linux_arm64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_linux_arm64
	# Compress MacOS servers
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_darwin_amd64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_darwin_amd64
	tar -czf $(SERVER_BUILD_DIR)/convoC2_server_darwin_arm64.tar.gz -C $(SERVER_BUILD_DIR) convoC2_server_darwin_arm64
	@echo "Compression complete"

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build directories"
