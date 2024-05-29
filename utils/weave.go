package weave

import (
	"errors"
	"os/exec"
	"regexp"
)

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
