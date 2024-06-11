package utils

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

func splitAndRemoveRightPart(input string) string {
	parts := strings.Split(input, "/")
	if len(parts) > 0 {
		return parts[0] // Return the part before the "/"
	}
	return "" // Return empty string if input does not contain "/"
}
func ListWeaveContainers() (map[string]string, error) {
	out, err := exec.Command("weave", "ps").Output()
	if err != nil {
		return nil, err
	}

	containerIPs := make(map[string]string)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			// Assuming the second part is the container ID and the third part is the IP address
			containerIPs[parts[0]] = splitAndRemoveRightPart(parts[2])
		}
	}
	return containerIPs, nil
}

// AttachContainer attaches a container to the Weave network.
func AttachContainer(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("containerID is empty")
	}

	out, err := exec.Command("weave", "attach", containerID).Output()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// FindIPAddressByContainerID finds the IP address associated with a container ID in the Weave network.
func FindIPAddressByContainerID(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("containerID is empty")
	}

	out, err := exec.Command("weave", "ps", containerID).Output()
	if err != nil {
		return "", err
	}

	// Convert byte array to string
	output := string(out)
	// Find IP address in the output string
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 2 {
		return "", errors.New("could not find IP address in command output")
	}

	// Return the first matched IP address
	return matches[1], nil
}
