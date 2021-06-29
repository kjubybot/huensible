package main

import (
	"net"
	"strings"
)

func parseIps(host string) []string {
	ips := make([]string, 0)
	if i := strings.IndexRune(host, '-'); i != -1 {
		start, end := ipToInt(net.ParseIP(host[:i]).To4()), ipToInt(net.ParseIP(host[i+1:]).To4())
		for i := start; i <= end; i++ {
			ips = append(ips, intToIp(i).String())
		}
	} else {
		ips = append(ips, host)
	}
	return ips
}

func ipToInt(ip net.IP) int32 {
	return int32(ip[0])<<24 | int32(ip[1])<<16 | int32(ip[2])<<8 | int32(ip[3])
}

func intToIp(i int32) net.IP {
	return net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}
