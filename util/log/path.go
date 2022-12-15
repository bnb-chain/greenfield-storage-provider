package log

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// StandardizePath is meant to decorate given file path by nodeIP/localIP/serviceName so that
// the returned path is consistent and unified throughout all services under node-real org
// The path after decoration will be `<file_path>/<node_ip>/<local_ip>/<service_name>.log`
func StandardizePath(filePath, serviceName string) string {
	return filepath.Join(filePath, os.Getenv("NODE_IP"), getLocalIP(), fmt.Sprintf("%s.log", serviceName))
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
