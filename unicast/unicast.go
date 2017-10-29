package unicast

import (
	"log"
	"net"
)

const (
	maxDatagramSize = 8192
)

// Listen binds to the TCP address and port given and writes packets received
// from that address to a buffer which is passed to a hander
func Listen(address string, handler func(int, []byte, string) []byte) {
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
		buffer := make([]byte, maxDatagramSize)
		numBytes, err := conn.Read(buffer)
		if err != nil {
			log.Println("ReadFromTCP failed:", err)
		}

		conn.Write(handler(numBytes, buffer, conn.RemoteAddr().String()))
	}
}

// NewSender creates a new TCP connection on which to send msg
func NewSender(address string) (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
