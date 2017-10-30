package caste

import (
	"fmt"
	"log"
	"net"

	"github.com/ccfortier/go/multicast"
	"github.com/ccfortier/go/unicast"
)

const (
	defaultMulticastIP     = "239.0.0.0"
	defaultMulticastPort   = 9000
	defaultUnicastIP       = "0.0.0.0"
	defaultCoordinatorPort = 10000
	defaultNetwork         = "172.17.0"
	//defaultBroadcastAddres = "172.17.255.255:9000"
	maxDatagramSize = 8192
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
	joinCaste(&cp)
	if cp.Coordinator == cp.PId {
		startAsCoordinator(&cp)
	} else {
		startAsWorker(&cp)
	}
}

func (cp CasteProcess) Dump() {
	log.Printf("(P:%d) Dump: %+v\n", cp.PId, cp)
}

func (cp CasteProcess) CheckCoordinator() {
	if cp.OnElection {
		log.Printf("(P:%d) Can't check coordinator. On election!\n", cp.PId, cp.Coordinator)
		return
	}
	ip := coordinatorAddress(&cp)
	if !unicastSendMessage(&cp, ip, "Are you ok?") {
		log.Printf("(P:%d) Coordinator [P:%d] is down!\n", cp.PId, cp.Coordinator)
		cp.BecomeCoordinator()
	}
}

func (cp CasteProcess) StartElection() {
	if cp.CId == cp.HCId {
		cp.BecomeCoordinator()
	}
}

func (cp CasteProcess) BecomeCoordinator() {
	log.Printf("(P:%d) I am the new coordinator\n", cp.PId)
	cp.Coordinator = cp.CId
	cp.Start()
}

func startAsCoordinator(cp *CasteProcess) {
	ip := coordinatorAddress(cp)
	log.Printf("(P:%d) Listen broadcast at %s...\n", cp.PId, broadcastAddress())
	go broadcastListen(cp)
	log.Printf("(P:%d) Started as coordinator. Waiting for requests at %s...\n", cp.PId, ip)
	go unicastListen(cp, ip)
}

func startAsWorker(cp *CasteProcess) {
	log.Printf("(P:%d) Listen broadcast at %s...\n", cp.PId, broadcastAddress())
	go broadcastListen(cp)
	log.Printf("(P:%d) Started as worker. Looking for coordinator at %s\n", cp.PId, coordinatorAddress(cp))
}

func joinCaste(cp *CasteProcess) {
	log.Printf("(P:%d) Joined caste %d. listen multicast at %s\n", cp.PId, cp.CId, casteAddress(cp))
	go multicastListen(cp)
}

func coordinatorAddress(cp *CasteProcess) string {
	if cp.SingleIP == 0 {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.Coordinator, defaultCoordinatorPort)
	} else {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.SingleIP, defaultCoordinatorPort+cp.Coordinator)
	}
}

func casteAddress(cp *CasteProcess) string {
	return fmt.Sprintf("%s:%d", defaultMulticastIP, defaultMulticastPort+cp.CId)
}

func broadcastAddress() string {
	return fmt.Sprintf("%s:%d", defaultMulticastIP, defaultMulticastPort)
}

func unicastListen(cp *CasteProcess, addr string) {

	unicast.Listen(addr, cp.unicastMsgHandlerOK)
}

func (cp CasteProcess) unicastMsgHandlerOK(n int, b []byte, addr string) []byte {
	log.Printf("(P:%d) Message received from %s: %s\n", cp.PId, addr, string(b[:n]))
	return []byte("ok")
}

func unicastSendMessage(cp *CasteProcess, addr string, msg string) bool {
	conn, err := unicast.NewSender(addr)
	if err != nil {
		//log.Printf("(P:%d) %s, when try ucast for %d at %s.\n", cp.PId, err, cp.Coordinator, addr)
		return false
	} else {
		defer conn.Close()

		conn.Write([]byte(fmt.Sprintf("[P:%d] %s", cp.PId, msg)))

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

func multicastListen(cp *CasteProcess) {

	multicast.Listen(casteAddress(cp), cp.multicastMsgHandlerOK)
}

func broadcastListen(cp *CasteProcess) {
	multicast.Listen(broadcastAddress(), cp.multicastMsgHandlerOK)
}

func (cp CasteProcess) multicastMsgHandlerOK(src *net.UDPAddr, n int, b []byte) []byte {
	log.Printf("(P:%d) Message received from %s: %s\n", cp.PId, src.String(), string(b[:n]))
	return []byte("ok")
}
