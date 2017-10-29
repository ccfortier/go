package caste

import (
	"log"
	"time"

	"github.com/ccfortier/go/unicast"
)

const (
	defaultMulticastAddress = "239.0.0.0"
	defaultUnicastAddress   = "0.0.0.0"
	defaultUnicastPort      = "10000"
	defaultNetwork          = "172.17.0."
	maxDatagramSize         = 8192
)

type CasteProcess struct {
	PId         int
	CId         int
	HCId        int
	Coordinator int
	OnElection  bool
}

func (cp CasteProcess) Start() {
	if cp.Coordinator == cp.PId {
		go startAsCoordinator(&cp)
	} else {
		go startAsNormal(&cp)
	}
}

func startAsCoordinator(cp *CasteProcess) {
	log.Printf("Process %d started as coordinator. Waiting for requests...", cp.PId)
	listen()
}

func startAsNormal(cp *CasteProcess) {
	log.Printf("Process %d started as normal. Waiting for requests...", cp.PId)
	for {
		ip := defaultNetwork + string(cp.Coordinator) + ":" + defaultUnicastPort
		if !checkProcess(ip) {
			log.Println("Coordinator is down!")
		}
		time.Sleep(1 * time.Second)
	}
}

func listen() {
	unicast.Listen(defaultUnicastAddress+":"+defaultUnicastPort, msgHandlerOK)
}

func msgHandlerOK(n int, b []byte) string {
	log.Println(string(b[:n]))
	return "ok"
}

func checkProcess(addr string) bool {
	conn, err := unicast.NewSender(addr)
	if err != nil {
		log.Printf("%s, when try ucast for %s.\n", err, addr)
		return false
	} else {
		defer conn.Close()

		conn.Write([]byte("Are you ok?"))

		buffer := make([]byte, maxDatagramSize)
		_, err := conn.Read(buffer)
		if err != nil {
			log.Println("ReadFromTCP failed to colect response:", err)
			return false
		}

		return true
	}
}
