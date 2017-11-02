package multicast

import (
	"log"
	"net"
)

const (
	maxDatagramSize = 8192
)

// Listen binds to the UDP address and port given and writes packets received
// from that address to a buffer which is passed to a hander
func Listen(conn *net.UDPConn, handler func(*net.UDPAddr, int, []byte), stop chan bool) {
	conn.SetReadBuffer(maxDatagramSize)
	// Loop forever reading from the socket
	for {
		buffer := make([]byte, maxDatagramSize)
		numBytes, src, err := conn.ReadFromUDP(buffer)
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
			log.Println("ReadFromUDP failed: ", err)
		} else {
			handler(src, numBytes, buffer)
		}

	}
}

func NewListener(address string) (*net.UDPConn, error) {
	// Parse the string address
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	// Open up a connection
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// NewSender creates a new UDP multicast connection on which to send msg
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
