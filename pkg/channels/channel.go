package channels

import (
	"fmt"
	"sync"
)

// Channel represents a communication channel for C2 traffic
type Channel interface {
	// Name returns the channel name
	Name() string

	// Description returns a human-readable description
	Description() string

	// Initialize prepares the channel for use
	Initialize(config map[string]string) error

	// SendCommand sends a command to an agent
	SendCommand(agentID, command string) error

	// IsAvailable checks if the channel is currently available
	IsAvailable() bool

	// Priority returns the channel priority (higher is tried first)
	Priority() int

	// GetConfig returns the channel configuration
	GetConfig() map[string]string
}

// ChannelRegistry manages available communication channels
type ChannelRegistry struct {
	channels map[string]Channel
	mutex    sync.RWMutex
}

// NewChannelRegistry creates a new channel registry
func NewChannelRegistry() *ChannelRegistry {
	return &ChannelRegistry{
		channels: make(map[string]Channel),
	}
}

// RegisterChannel adds a channel to the registry
func (cr *ChannelRegistry) RegisterChannel(channel Channel) error {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	name := channel.Name()
	if _, exists := cr.channels[name]; exists {
		return fmt.Errorf("channel %s already registered", name)
	}

	cr.channels[name] = channel
	return nil
}

// GetChannel retrieves a channel by name
func (cr *ChannelRegistry) GetChannel(name string) (Channel, error) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	channel, exists := cr.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel %s not found", name)
	}

	return channel, nil
}

// GetAvailableChannels returns all available channels sorted by priority
func (cr *ChannelRegistry) GetAvailableChannels() []Channel {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	// Collect available channels
	available := make([]Channel, 0)
	for _, channel := range cr.channels {
		if channel.IsAvailable() {
			available = append(available, channel)
		}
	}

	// Sort by priority (higher first)
	for i := 0; i < len(available); i++ {
		for j := i + 1; j < len(available); j++ {
			if available[i].Priority() < available[j].Priority() {
				available[i], available[j] = available[j], available[i]
			}
		}
	}

	return available
}

// BaseChannel implements common functionality for channels
type BaseChannel struct {
	name        string
	description string
	priority    int
	available   bool
	config      map[string]string
	mutex       sync.RWMutex
}

// NewBaseChannel creates a new base channel
func NewBaseChannel(name, description string, priority int) *BaseChannel {
	return &BaseChannel{
		name:        name,
		description: description,
		priority:    priority,
		available:   false,
		config:      make(map[string]string),
	}
}

// Name returns the channel name
func (bc *BaseChannel) Name() string {
	return bc.name
}

// Description returns the channel description
func (bc *BaseChannel) Description() string {
	return bc.description
}

// Priority returns the channel priority
func (bc *BaseChannel) Priority() int {
	return bc.priority
}

// IsAvailable checks if the channel is available
func (bc *BaseChannel) IsAvailable() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return bc.available
}

// SetAvailable updates the channel availability
func (bc *BaseChannel) SetAvailable(available bool) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	bc.available = available
}

// SetConfig sets the channel configuration
func (bc *BaseChannel) SetConfig(config map[string]string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.config = make(map[string]string)
	for k, v := range config {
		bc.config[k] = v
	}
}

// GetConfig returns the channel configuration
func (bc *BaseChannel) GetConfig() map[string]string {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	config := make(map[string]string)
	for k, v := range bc.config {
		config[k] = v
	}

	return config
}

// GetConfigValue retrieves a configuration value
func (bc *BaseChannel) GetConfigValue(key string) (string, bool) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	value, exists := bc.config[key]
	return value, exists
}

// SetConfigValue sets a configuration value
func (bc *BaseChannel) SetConfigValue(key, value string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.config[key] = value
}
