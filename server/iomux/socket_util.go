package iomux

import (
	"net"
	"syscall"
)

func socketFD(conn net.Conn) int {
	if con, ok := conn.(syscall.Conn); ok {
		raw, err := con.SyscallConn()
		if err != nil {
			return 0
		}
		sfd := 0
		raw.Control(func(fd uintptr) {
			sfd = int(fd)
		})
		return sfd
	}
	return 0
}

/**
Converts ipv4 address to an array of bytes
Eg:
	I/P : 127.0.0.1
	o/p: [127,0,0,1] as byte array
*/
func convertIPV4StrtoArray(hostIpv4 string) [4]byte {
	ip4Bytes := net.ParseIP(hostIpv4)
	return [4]byte{ip4Bytes[0], ip4Bytes[1], ip4Bytes[2], ip4Bytes[3]}
}
