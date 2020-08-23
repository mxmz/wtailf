package util

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type InterfaceAddress struct {
	IP  net.IP
	Net *net.IPNet
}

// GetNetInterfaceAddresses returns networks interfaces skipping "lo", "docker" and VPNs
func GetNetInterfaceAddresses() []InterfaceAddress {

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error, no interfaces: %s\n", err)
		return []InterfaceAddress{}
	}
	var rv = []InterfaceAddress{}
IFACES:
	for _, iface := range interfaces {
		fmt.Printf("%s\n", iface.Name)
		for _, toSkip := range []string{"lo", "docker", "tun", "vpn"} {
			if strings.HasPrefix(iface.Name, toSkip) {
				continue IFACES
			}
		}

		addrs, err := iface.Addrs()

		if err != nil {
			log.Printf(" %s. %s\n", iface.Name, err)
			continue
		}
		for _, a := range addrs {
			addr := a.String()
			if !strings.Contains(addr, ":") {
				ip, ipnet, err := net.ParseCIDR(addr)
				fmt.Printf("%s %s\n", ip, ipnet)
				if err == nil {
					rv = append(rv, InterfaceAddress{ip, ipnet})
				}
			}
		}
	}
	return rv
}
