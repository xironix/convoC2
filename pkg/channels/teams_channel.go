package channels

import (
	"fmt"

	"github.com/cxnturi0n/convoC2/pkg/crypto"
)

// TeamsChannel represents the Microsoft Teams communication channel
type TeamsChannel struct {
	*BaseChannel
}

// NewTeamsChannel creates a new Teams channel
func NewTeamsChannel() *TeamsChannel {
	base := NewBaseChannel(
		"teams",
		"Microsoft Teams communication channel",
		100, // Highest priority
	)

	return &TeamsChannel{
		BaseChannel: base,
	}
}

// Initialize prepares the Teams channel for use
func (tc *TeamsChannel) Initialize(config map[string]string) error {
	tc.SetConfig(config)
	tc.SetAvailable(true)
	return nil
}

// SendCommand sends a command to an agent
func (tc *TeamsChannel) SendCommand(agentID, command string) error {
	// Encrypt the command
	encryptedCommand, err := crypto.Encrypt([]byte(command), crypto.DefaultKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt command: %w", err)
	}

	// We would send the command here in a real implementation
	fmt.Printf("Teams channel sending command to agent %s: %s\n", agentID, encryptedCommand)

	return nil
}
