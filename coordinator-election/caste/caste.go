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
)

type CasteProcess struct {
	PId               int
	CId               int
	HCId              int
	Coordinator       int
	OnElection        bool
	SingleIP          int
	StopChanel        chan bool
	UnicastListener   *net.TCPListener
	MulticastListener *net.UDPConn
	BroadcastListener *net.UDPConn
}

func (cp CasteProcess) Start() (*net.TCPListener, *net.UDPConn, *net.UDPConn) {
	var startAs string
	if cp.Coordinator == cp.PId {
		startAs = "COORDINATOR"
	} else {
		startAs = "WORKER"
	}
	ip := processAddress(&cp)
	log.Printf("(P:%d) Started as %s. Waiting for requests at %s...\n", cp.PId, startAs, ip)
	unicastListen(&cp)
	log.Printf("(P:%d) Joined caste %d. listen multicast at %s...\n", cp.PId, cp.CId, casteAddress(&cp))
	multicastListen(&cp)
	log.Printf("(P:%d) Listen broadcast at %s...\n", cp.PId, broadcastAddress())
	broadcastListen(&cp)
	return cp.UnicastListener, cp.MulticastListener, cp.BroadcastListener
}

func (cp CasteProcess) Dump() {
	log.Printf("(P:%d) Dump: %+v\n", cp.PId, cp)
}

func (cp CasteProcess) Encode() string {
	return fmt.Sprintf("%v", cp)
}

func (cp CasteProcess) Decode(cpEncoded string) {
	fmt.Sscanf(cpEncoded[:len(cpEncoded)-1], "{%d %d %d %d %t %d}", &cp.PId, &cp.CId, &cp.HCId, &cp.Coordinator, &cp.OnElection, &cp.SingleIP)
}

func (cp CasteProcess) CasteMsg(msg string) *CasteMsg {
	return &CasteMsg{cp.PId, cp.CId, msg}
}

func (cp CasteProcess) StopListen() {
	cp.StopChanel <- true
	cp.UnicastListener.Close()

	cp.StopChanel <- true
	cp.MulticastListener.Close()

	cp.StopChanel <- true
	cp.BroadcastListener.Close()
}

func (cp CasteProcess) CheckCoordinator() {
	if cp.OnElection {
		log.Printf("(P:%d) Can't check coordinator. On election!\n", cp.PId, cp.Coordinator)
		return
	}
	ip := coordinatorAddress(&cp)
	_, err := unicastSendMessage(&cp, ip, cp.CasteMsg("AreYouOk?"))
	if err != nil {
		log.Printf("(P:%d) Coordinator [P:%d] is down!\n", cp.PId, cp.Coordinator)
		cp.StartElection()
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
	cp.StopListen()
	time.Sleep(10 * time.Millisecond)
	cp.Start()
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

func casteAddress(cp *CasteProcess) string {
	return fmt.Sprintf("%s:%d", defaultMulticastIP, defaultMulticastPort+cp.CId)
}

func broadcastAddress() string {
	return fmt.Sprintf("%s:%d", defaultMulticastIP, defaultMulticastPort)
}

func unicastListen(cp *CasteProcess) {
	l, err := unicast.NewListener(processAddress(cp))
	cp.UnicastListener = l
	if err == nil {
		go unicast.Listen(cp.UnicastListener, cp.unicastMsgHandler, cp.StopChanel)
	}
}

func (cp CasteProcess) unicastMsgHandler(n int, b []byte, addr string) []byte {
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
		if msg.Text != "STOP" {
			conn.Write([]byte(msg.Encode()))
		} else {
			conn.Write([]byte(fmt.Sprintf("STOP%d", cp.PId)))
		}

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
	l, err := multicast.NewListener(casteAddress(cp))
	cp.MulticastListener = l
	if err == nil {
		go multicast.Listen(cp.MulticastListener, cp.multicastMsgHandler, cp.StopChanel)
	}
}

func broadcastListen(cp *CasteProcess) {
	l, err := multicast.NewListener(broadcastAddress())
	cp.BroadcastListener = l
	if err == nil {
		go multicast.Listen(cp.BroadcastListener, cp.multicastMsgHandler, cp.StopChanel)
	}
}

func (cp CasteProcess) multicastMsgHandler(src *net.UDPAddr, n int, b []byte) []byte {
	var msg CasteMsg
	var returnMsg *CasteMsg
	msg.Decode(string(b[:n]))
	switch msg.Text {
	case "ok!":
		returnMsg = cp.CasteMsg("XXX")
	default:
		returnMsg = cp.CasteMsg("ok...")
	}
	log.Printf("(P:%d) Message received from %s: %s\n", cp.PId, src.String(), string(b[:n]))
	return []byte(returnMsg.Encode())
}

func multicastSendMessage(cp *CasteProcess, addr string, msg *CasteMsg) error {
	conn, err := multicast.NewSender(addr)
	if err != nil {
		return err
	} else {
		if msg.Text != "STOP" {
			conn.Write([]byte(msg.Encode()))
		} else {
			conn.Write([]byte(fmt.Sprintf("STOP%d", cp.PId)))
		}
		return nil
	}
}
