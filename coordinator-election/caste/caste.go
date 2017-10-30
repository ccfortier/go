package caste

import (
	"fmt"
	"log"

	"github.com/ccfortier/go/unicast"
)

const (
	defaultMulticastAddress = "239.0.0.0"
	defaultUnicastAddress   = "0.0.0.0"
	defaultCoordinatorPort  = 10000
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
		go startAsWorker(&cp)
	}
}

func (cp CasteProcess) CheckCoordinator() {
	if cp.OnElection {
		log.Printf("(P:%d) Can't check coordinator. On election!\n", cp.PId, cp.Coordinator)
		return
	}
	ip := coordinatorIP(&cp)
	if !checkProcess(&cp, ip) {
		log.Printf("(P:%d) Coordinator [P:%d] is down!\n", cp.PId, cp.Coordinator)
	}
}

func (cp CasteProcess) Dump() {
	log.Printf("(P:%d) Dump: %+v\n", cp.PId, cp)
}

func startAsCoordinator(cp *CasteProcess) {
	ip := coordinatorIP(cp)
	log.Printf("(P:%d) started as coordinator. Waiting for requests at %s...", cp.PId, ip)
	listen(cp, ip)
}

func startAsWorker(cp *CasteProcess) {
	log.Printf("(P:%d) started as worker. Looking for coordinator at %s", cp.PId, coordinatorIP(cp))
	cp.CheckCoordinator()
}

func coordinatorIP(cp *CasteProcess) string {
	if cp.SingleIP == 0 {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.Coordinator, defaultCoordinatorPort)
	} else {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.SingleIP, defaultCoordinatorPort+cp.Coordinator)
	}
}

func listen(cp *CasteProcess, addr string) {

	unicast.Listen(addr, cp.msgHandlerOK)
}

func (cp CasteProcess) msgHandlerOK(n int, b []byte, addr string) []byte {
	log.Printf("(P:%d) Message received from %s: %s\n", cp.PId, addr, string(b[:n]))
	return []byte("ok")
}

func checkProcess(cp *CasteProcess, addr string) bool {
	conn, err := unicast.NewSender(addr)
	if err != nil {
		//log.Printf("(P:%d) %s, when try ucast for %d at %s.\n", cp.PId, err, cp.Coordinator, addr)
		return false
	} else {
		defer conn.Close()

		conn.Write([]byte(fmt.Sprintf("[P:%d] Are you ok?", cp.PId)))

		buffer := make([]byte, maxDatagramSize)
		n, err := conn.Read(buffer)
		if err != nil {
			//log.Printf("(P:%d) ReadFromTCP failed to colect response: %s", err)
			return false
		}

		response := fmt.Sprintf("%s", buffer[:n])
		log.Printf("(P:%d) Response from coordinator [P:%d]: %s", cp.PId, cp.Coordinator, response)
		return response == "ok"
	}
}
