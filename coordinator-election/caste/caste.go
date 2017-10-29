package caste

import (
	"fmt"
	"log"

	"github.com/ccfortier/go/unicast"
)

const (
	defaultMulticastAddress = "239.0.0.0"
	defaultUnicastAddress   = "0.0.0.0"
	defaultUnicastPort      = 10000
	defaultNetwork          = "172.17.0"
	maxDatagramSize         = 8192
)

type CasteProcess struct {
	PId         int
	CId         int
	HCId        int
	Coordinator int
	OnElection  bool
	SingleIP    int
}

func (cp CasteProcess) Start() {
	if cp.Coordinator == cp.PId {
		go startAsCoordinator(&cp)
	} else {
		go startAsNormal(&cp)
	}
}

func (cp CasteProcess) CheckCoordinator() {
	ip := coordinatorIP(&cp)
	if !checkProcess(ip, cp.PId) {
		log.Println("Coordinator is down!")
	}
}

func startAsCoordinator(cp *CasteProcess) {
	ip := processIP(cp)
	log.Printf("Process %d started as coordinator. Waiting for requests at %s...", cp.PId, ip)
	listen(ip)
}

func startAsNormal(cp *CasteProcess) {
	log.Printf("Process %d started as normal. Looking for coordinator at %s", cp.PId, coordinatorIP(cp))
	cp.CheckCoordinator()
}

func processIP(cp *CasteProcess) string {
	if cp.SingleIP == 0 {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.PId, defaultUnicastPort)
	} else {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.SingleIP, defaultUnicastPort+cp.PId)
	}
}

func coordinatorIP(cp *CasteProcess) string {
	if cp.SingleIP == 0 {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.Coordinator, defaultUnicastPort)
	} else {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.SingleIP, defaultUnicastPort+cp.Coordinator)
	}
}

func listen(addr string) {

	unicast.Listen(addr, msgHandlerOK)
}

func msgHandlerOK(n int, b []byte, addr string) []byte {
	log.Printf("Message received from %s: %s\n", addr, string(b[:n]))
	return []byte("ok")
}

func checkProcess(addr string, pID int) bool {
	conn, err := unicast.NewSender(addr)
	if err != nil {
		log.Printf("%s, when try ucast for %s.\n", err, addr)
		return false
	} else {
		defer conn.Close()

		conn.Write([]byte(fmt.Sprintf("[Process:%d] Are you ok?", pID)))

		buffer := make([]byte, maxDatagramSize)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("ReadFromTCP failed to colect response:", err)
			return false
		}

		response := fmt.Sprintf("%s", buffer[:n])
		log.Printf("Response from coordinator: %s", response)
		return response == "ok"
	}
}
