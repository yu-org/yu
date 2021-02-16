package ip

import "strings"

func MakePort(port string) string {
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func MakeLocalIp(port string) string {
	if strings.HasPrefix(port, ":") {
		return "localhost" + port
	}
	return "localhost:" + port
}

func MakeIp(host, port string) string {
	if strings.HasPrefix(port, ":") {
		return host + port
	}
	return host + ":" + port
}
