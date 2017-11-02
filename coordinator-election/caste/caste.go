package caste

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ccfortier/go/multicast"
	"github.com/ccfortier/go/unicast"
)

const (
	defaultMulticastIP   = "239.0.0.0"
	defaultMulticastPort = 9000
	defaultUnicastIP     = "0.0.0.0"
	defaultUnicastPort   = 10000
	defaultNetwork       = "172.17.0"
	maxDatagramSize      = 8192
	defaultTimeOut       = 10
)

type CasteProcess struct {
	PId               int
	CId               int
	HCId              int
	Coordinator       int
	Status            string
	SingleIP          int
	StopChan          chan bool
	CandidateChan     chan int
	UnicastListener   *net.TCPListener
	MulticastListener *net.UDPConn
	BroadcastListener *net.UDPConn
}

func (cp *CasteProcess) Start() (*net.TCPListener, *net.UDPConn, *net.UDPConn) {
	var startAs string
	if cp.Coordinator == cp.PId {
		startAs = "COORDINATOR"
	} else {
		startAs = "WORKER"
	}
	ip := processAddress(cp)
	log.Printf("(P:%d) Started as %s. Waiting for requests at %s...\n", cp.PId, startAs, ip)
	unicastListen(cp)
	log.Printf("(P:%d) Joined caste %d. listen multicast at %s...\n", cp.PId, cp.CId, casteAddress(cp.CId))
	multicastListen(cp)
	log.Printf("(P:%d) Listen broadcast at %s...\n", cp.PId, broadcastAddress())
	broadcastListen(cp)
	return cp.UnicastListener, cp.MulticastListener, cp.BroadcastListener
}

func (cp *CasteProcess) Dump() {
	log.Printf("(P:%d) Dump: {PId=%d CId=%d HCId=%d Coordinator=%d Status=%s}", cp.PId, cp.PId, cp.CId, cp.HCId, cp.Coordinator, cp.Status)
}

func (cp *CasteProcess) Encode() string {
	return fmt.Sprintf("{%d %d %d %d %s}", cp.PId, cp.CId, cp.HCId, cp.Coordinator, cp.Status)
}

func (cp *CasteProcess) Decode(cpEncoded string) {
	fmt.Sscanf(cpEncoded[:len(cpEncoded)-1], "{%d %d %d %d %s}", cp.PId, cp.CId, cp.HCId, cp.Coordinator, cp.Status)
}

func (cp *CasteProcess) CasteMsg(msg string) *CasteMsg {
	return &CasteMsg{cp.PId, cp.CId, msg}
}

func (cp *CasteProcess) StopListen() {
	cp.StopChan <- true
	cp.UnicastListener.Close()

	cp.StopChan <- true
	cp.MulticastListener.Close()

	cp.StopChan <- true
	cp.BroadcastListener.Close()

	log.Printf("(P:%d) Stopped listening.", cp.PId)
}

func (cp *CasteProcess) CheckCoordinator() {
	if cp.Status == "WaitingElection" {
		log.Printf("(P:%d) Can't check coordinator. Waiting election!\n", cp.PId)
	} else {
		ip := coordinatorAddress(cp)
		_, err := unicastSendMessage(cp, ip, cp.CasteMsg("AreYouOk?"))
		if err != nil {
			log.Printf("(P:%d) Coordinator [P:%d] is down!\n", cp.PId, cp.Coordinator)
			cp.StartElection()
		}
	}
}

func (cp *CasteProcess) StartElection() {
	if cp.CId == cp.HCId {
		cp.BecomeCoordinator()
	} else {
		multicastSendMessage(cp, casteAddress(cp.CId), cp.CasteMsg("WaitElection!"))
		//multicastSendMessage(cp, casteAddress(cp.CId+1), cp.CasteMsg("SomeoneUp?"))
		cp.Status = "WaitingCandidate"
		go func() {
			select {
			case candidate := <-cp.CandidateChan:
				log.Printf("Candidate: [P:%d]", candidate)
			case <-time.After(time.Millisecond * defaultTimeOut):
				log.Printf("(P:%d) No candidates in superior castes", cp.PId)
				cp.BecomeCoordinator()
			}
		}()
	}
}

func (cp *CasteProcess) BecomeCoordinator() {
	log.Printf("(P:%d) I am the new coordinator\n", cp.PId)
	cp.Coordinator = cp.PId
	cp.StopListen()
	time.Sleep(10 * time.Millisecond)
	cp.Start()
	multicastSendMessage(cp, broadcastAddress(), cp.CasteMsg("IAmCoordinator!"))
}

func processAddress(cp *CasteProcess) string {
	if cp.SingleIP == 0 {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.PId, defaultUnicastPort)
	} else {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.SingleIP, defaultUnicastPort+cp.PId)
	}
}

func coordinatorAddress(cp *CasteProcess) string {
	if cp.SingleIP == 0 {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.Coordinator, defaultUnicastPort)
	} else {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, cp.SingleIP, defaultUnicastPort+cp.Coordinator)
	}
}

func casteAddress(caste int) string {
	return fmt.Sprintf("%s:%d", defaultMulticastIP, defaultMulticastPort+caste)
}

func broadcastAddress() string {
	return fmt.Sprintf("%s:%d", defaultMulticastIP, defaultMulticastPort)
}

func unicastListen(cp *CasteProcess) {
	l, err := unicast.NewListener(processAddress(cp))
	cp.UnicastListener = l
	if err == nil {
		go unicast.Listen(cp.UnicastListener, cp.unicastMsgHandler, cp.StopChan)
	}
}

func (cp *CasteProcess) unicastMsgHandler(n int, b []byte, addr string) []byte {
	var msg CasteMsg
	var returnMsg *CasteMsg
	msg.Decode(string(b[:n]))
	switch msg.Text {
	case "AreYouOk?":
		returnMsg = cp.CasteMsg("ok!")
	default:
		returnMsg = cp.CasteMsg("ok...")
	}
	log.Printf("(P:%d) Message received from [P:%d] at %s: %s\n", cp.PId, msg.PId, addr, msg.Text)
	return []byte(returnMsg.Encode())
}

func unicastSendMessage(cp *CasteProcess, addr string, msg *CasteMsg) (*CasteMsg, error) {
	var response CasteMsg
	conn, err := unicast.NewSender(addr)
	if err != nil {
		return nil, err
	} else {
		conn.Write([]byte(msg.Encode()))
		buffer := make([]byte, maxDatagramSize)
		n, err := conn.Read(buffer)
		if err != nil {
			return nil, fmt.Errorf("(P:%d) ReadFromTCP failed to colect response: %s", cp.PId, err)
		}
		if string(buffer)[:4] != "STOP" {
			response.Decode(string(buffer[:n]))
			log.Printf("(P:%d) Response from process [P:%d]: %s", cp.PId, cp.Coordinator, response.Text)
			return &response, nil
		} else {
			return nil, nil
		}
	}
}

func multicastListen(cp *CasteProcess) {
	l, err := multicast.NewListener(casteAddress(cp.CId))
	cp.MulticastListener = l
	if err == nil {
		go multicast.Listen(cp.MulticastListener, cp.multicastMsgHandler, cp.StopChan)
	}
}

func broadcastListen(cp *CasteProcess) {
	l, err := multicast.NewListener(broadcastAddress())
	cp.BroadcastListener = l
	if err == nil {
		go multicast.Listen(cp.BroadcastListener, cp.multicastMsgHandler, cp.StopChan)
	}
}

func (cp *CasteProcess) multicastMsgHandler(src *net.UDPAddr, n int, b []byte) {
	var msg CasteMsg
	msg.Decode(string(b[:n]))
	if cp.PId != msg.PId {
		log.Printf("(P:%d) Message received from [P:%d] at %s: %s\n", cp.PId, msg.PId, src.String(), msg.Text)
		switch msg.Text {
		case "IAmCoordinator!":
			cp.Coordinator = msg.PId
			cp.HCId = msg.CId
			cp.Status = "Up"
		case "WaitElection!":
			cp.Status = "WaitingElection"
		default:
		}
	}
}

func multicastSendMessage(cp *CasteProcess, addr string, msg *CasteMsg) error {
	conn, err := multicast.NewSender(addr)
	if err != nil {
		return err
	} else {
		conn.Write([]byte(msg.Encode()))
		return nil
	}
}
