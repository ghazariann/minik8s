package kubeproxy

import (
	"fmt"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"os"
	"strings"
)

type dnsManager struct {
}

func FormatNginxConfig(dns apiobject.Dns) string {
	var sb strings.Builder

	// Add comment and server block start
	sb.WriteString(fmt.Sprintf("# %s.conf\n", dns.Spec.Hostname))
	sb.WriteString(fmt.Sprintf("server {\n\tlisten %d;\n\tserver_name %s;\n", 80, dns.Spec.Hostname))

	// Template for location blocks
	locationTemplate := "\tlocation %s {\n\t\tproxy_pass http://%s:%s/;\n\t}\n"

	// Add location blocks for each path
	for _, p := range dns.Spec.Paths {
		path := p.UrlPath
		if path[0] != '/' {
			path = "/" + path
		}
		sb.WriteString(fmt.Sprintf(locationTemplate, path, p.ServiceIp, p.ServicePort))
	}

	// Add server block end
	sb.WriteString("}")

	return sb.String()
}
func WriteNginxConfig(dns apiobject.Dns, conf string) error {
	configFileName := fmt.Sprintf("%s.conf", dns.Spec.Hostname)
	configFilePath := fmt.Sprintf("%s%s", configs.NginxConfigFolderPath, configFileName)

	// Check if the configuration directory exists
	if _, err := os.Stat(configs.NginxConfigFolderPath); err != nil {
		if os.IsNotExist(err) {
			log.Printf("Configuration directory does not exist, creating: %s\n", configs.NginxConfigFolderPath)
			if err := os.MkdirAll(configs.NginxConfigFolderPath, os.ModePerm); err != nil {
				log.Printf("Error creating configuration directory: %v\n", err)
				return err
			}
		} else {
			log.Printf("Error checking configuration directory: %v\n", err)
			return err
		}
	}

	// Create or open the configuration file
	file, err := os.Create(configFilePath)
	if err != nil {
		log.Printf("Error creating configuration file: %v\n", err)
		return err
	}
	defer file.Close()

	// Truncate the file to ensure it's empty
	if err := file.Truncate(0); err != nil {
		log.Printf("Error truncating configuration file: %v\n", err)
		return err
	}

	// Write the configuration string to the file
	if _, err := file.Write([]byte(conf)); err != nil {
		log.Printf("Error writing to configuration file: %v\n", err)
		return err
	}

	log.Printf("Successfully wrote Nginx configuration to %s\n", configFilePath)
	return nil
}
func (d *dnsManager) AddDns(dnsStore apiobject.DnsStore) error {

	nginxConfig := FormatNginxConfig(*dnsStore.ToDns())
	WriteNginxConfig(*dnsStore.ToDns(), nginxConfig)
	return nil
}

func (d *dnsManager) DeleteDns(dns apiobject.DnsStore) error {
	// Construct the configuration file path
	configFileName := fmt.Sprintf("%s.conf", dns.Spec.Hostname)
	configFilePath := fmt.Sprintf("%s%s", configs.NginxConfigFolderPath, configFileName)

	// Check if the configuration file exists
	if _, err := os.Stat(configFilePath); err != nil {
		if os.IsNotExist(err) {
			// File does not exist, no action needed
			log.Printf("Configuration file does not exist: %s\n", configFilePath)
			return nil
		} else {
			// Other errors
			log.Printf("Error checking configuration file: %v\n", err)
			return err
		}
	}

	// Attempt to delete the configuration file
	if err := os.Remove(configFilePath); err != nil {
		log.Printf("Error removing configuration file: %v\n", err)
		return err
	}

	log.Printf("Successfully deleted Nginx configuration file: %s\n", configFilePath)
	return nil

}
