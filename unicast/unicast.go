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
func Listen(listener *net.TCPListener, handler func(int, []byte, string) []byte, stop chan bool) {
	// Loop forever reading from the socket
	for {
		conn, err := listener.Accept()
		if err != nil {
			stopNow := false
			select {
			case stopNow = <-stop:
			default:
				stopNow = false
			}
			if stopNow {
				break
			}
			log.Println(err)
		}
		buffer := make([]byte, maxDatagramSize)
		numBytes, err := conn.Read(buffer)
		if err != nil {
			log.Println("ReadFromTCP failed:", err)
		}
		handled := handler(numBytes, buffer, conn.RemoteAddr().String())
		conn.Write(handled)
	}
}

// NewListener creates a new TCP listener
func NewListener(address string) (*net.TCPListener, error) {
	// Parse the string address
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	// Open up a connection
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return l, nil
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
