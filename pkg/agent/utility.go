package agent

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/user"
	"strings"
	"time"
)

func getAgentID() (string, error) {
	hostname, err := getHostname()
	if err != nil {
		return "", err
	}

	user, err := getCurrentUser()
	if err != nil {
		return "", err
	}

	mac, err := getMACAddress()
	combined := hostname + user
	if err == nil {
		combined += mac
	}

	hash := md5.New()
	hash.Write([]byte(combined))

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String(), nil
		}
	}

	return "", errors.New("no active network interface found")
}

func fileBytesToString(filePath string) (string, error) {
	logFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file at %s: %w", filePath, err)
	}
	defer logFile.Close()

	logFileBytes, err := io.ReadAll(logFile)
	if err != nil {
		return "", fmt.Errorf("unable to read file at %s: %w", filePath, err)
	}

	return string(logFileBytes), nil
}

func getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return hostname, nil
}

func getCurrentUserFull() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	return currentUser.Username, nil
}

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Split the username in case it's in the form HOST\username
	usernameParts := strings.Split(currentUser.Username, `\`)
	username := usernameParts[len(usernameParts)-1]

	return username, nil
}

func random() string {
	hash := md5.New()
	randomData := time.Now().String() + fmt.Sprintf("%d", rand.Int())
	hash.Write([]byte(randomData))
	return hex.EncodeToString(hash.Sum(nil))
}
