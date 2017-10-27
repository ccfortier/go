package unicast

import (
	"log"
	"net"
)

const (
	maxDatagramSize = 8192
)

// Listen binds to the UDP address and port given and writes packets received
// from that address to a buffer which is passed to a hander
func Listen(address string, handler func(*net.TCPAddr, int, []byte)) {
	// Parse the string address
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	// Open up a connection
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	// Loop forever reading from the socket
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		conn.SetReadBuffer(maxDatagramSize)
		buffer := make([]byte, maxDatagramSize)
		numBytes, src, err := conn.Read(buffer)
		if err != nil {
			log.Fatal("ReadFromTCP failed:", err)
		}

		handler(src, numBytes, buffer)
	}
}

// NewSender creates a new TCP connection on which to send msg
func NewSender(address string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil

}
