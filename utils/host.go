package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
)

func GetHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return hostName
}

func GetHostIp() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	for _, i := range interfaces {

		name := i.Name
		if name == "ens33" || name == "eth0" || name == "ens3" {
			addrs, err := i.Addrs()
			if err != nil {
				fmt.Println("Error:", err)
				return "", err
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}
	return "", errors.New("no interface or no named ens33/ens3/eth0")
}
