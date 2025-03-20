package modules

import (
	"fmt"
	"sync"
)

// Module represents a C2 module that can be loaded and run
type Module interface {
	// Name returns the module's name
	Name() string

	// Description returns a description of the module
	Description() string

	// Author returns the module author's name
	Author() string

	// Options returns the module's available options
	Options() map[string]string

	// SetOption sets a module option
	SetOption(key, value string) error

	// Run executes the module
	Run(agentID string) (string, error)
}

// ModuleRegistry manages available modules
type ModuleRegistry struct {
	modules map[string]Module
	mutex   sync.RWMutex
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[string]Module),
	}
}

// RegisterModule adds a module to the registry
func (mr *ModuleRegistry) RegisterModule(module Module) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	name := module.Name()
	if _, exists := mr.modules[name]; exists {
		return fmt.Errorf("module %s already registered", name)
	}

	mr.modules[name] = module
	return nil
}

// GetModule retrieves a module by name
func (mr *ModuleRegistry) GetModule(name string) (Module, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}

	return module, nil
}

// ListModules returns all registered modules
func (mr *ModuleRegistry) ListModules() []Module {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	modules := make([]Module, 0, len(mr.modules))
	for _, module := range mr.modules {
		modules = append(modules, module)
	}

	return modules
}

// BaseModule implements common functionality for modules
type BaseModule struct {
	name            string
	description     string
	author          string
	options         map[string]string
	optionHelp      map[string]string
	requiredOptions []string
	mutex           sync.RWMutex
}

// NewBaseModule creates a new base module
func NewBaseModule(name, description, author string) *BaseModule {
	return &BaseModule{
		name:            name,
		description:     description,
		author:          author,
		options:         make(map[string]string),
		optionHelp:      make(map[string]string),
		requiredOptions: []string{},
	}
}

// Name returns the module's name
func (bm *BaseModule) Name() string {
	return bm.name
}

// Description returns the module's description
func (bm *BaseModule) Description() string {
	return bm.description
}

// Author returns the module's author
func (bm *BaseModule) Author() string {
	return bm.author
}

// Options returns the module's options
func (bm *BaseModule) Options() map[string]string {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	options := make(map[string]string)
	for k, v := range bm.options {
		options[k] = v
	}

	return options
}

// SetOption sets a module option
func (bm *BaseModule) SetOption(key, value string) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if _, exists := bm.optionHelp[key]; !exists {
		return fmt.Errorf("unknown option: %s", key)
	}

	bm.options[key] = value
	return nil
}

// RegisterOption adds an option to the module
func (bm *BaseModule) RegisterOption(name, help, defaultValue string, required bool) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	bm.optionHelp[name] = help
	bm.options[name] = defaultValue

	if required {
		bm.requiredOptions = append(bm.requiredOptions, name)
	}
}

// ValidateOptions checks if all required options are set
func (bm *BaseModule) ValidateOptions() error {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	for _, option := range bm.requiredOptions {
		value, exists := bm.options[option]
		if !exists || value == "" {
			return fmt.Errorf("required option %s is not set", option)
		}
	}

	return nil
}

// GetOption gets an option value
func (bm *BaseModule) GetOption(name string) (string, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	value, exists := bm.options[name]
	if !exists {
		return "", fmt.Errorf("option %s not found", name)
	}

	return value, nil
}

// OptionHelp returns help text for options
func (bm *BaseModule) OptionHelp() map[string]string {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	help := make(map[string]string)
	for k, v := range bm.optionHelp {
		help[k] = v
	}

	return help
}
