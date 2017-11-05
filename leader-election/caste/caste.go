package caste

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	defaultMulticastIP   = "239.0.0.0"
	defaultMulticastPort = 9000
	defaultUnicastIP     = "0.0.0.0"
	defaultUnicastPort   = 15000
	defaultNetwork       = "172.17.0"
	maxDatagramSize      = 8192
	defaultTimeOut       = 10
)

type CasteProcess struct {
	PId               int
	CId               int
	HCId              int
	Leader            int
	Status            string
	Candidate         int
	SingleIP          int
	StopChan          chan bool
	CandidateChan     chan int
	UnicastListener   *net.TCPListener
	MulticastListener *net.UDPConn
	BroadcastListener *net.UDPConn
	FLog              *os.File
	QuietMode         *bool
	mux               sync.Mutex
}

func (cp *CasteProcess) Start() (*net.TCPListener, *net.UDPConn, *net.UDPConn) {
	unicastListen(cp)
	multicastListen(cp)
	broadcastListen(cp)
	return cp.UnicastListener, cp.MulticastListener, cp.BroadcastListener
}

func (cp *CasteProcess) Dump() {
	log.Printf("(P:%d-%d) {PId=%d CId=%d HCId=%d Leader=%d Status=%s}", cp.PId, cp.CId, cp.PId, cp.CId, cp.HCId, cp.Leader, cp.Status)
}

func (cp *CasteProcess) Encode() string {
	return fmt.Sprintf("{%d %d %d %d %s}", cp.PId, cp.CId, cp.HCId, cp.Leader, cp.Status)
}

func (cp *CasteProcess) Decode(cpEncoded string) {
	fmt.Sscanf(cpEncoded[:len(cpEncoded)-1], "{%d %d %d %d %s}", cp.PId, cp.CId, cp.HCId, cp.Leader, cp.Status)
}

func (cp *CasteProcess) Msg(msg string) *CasteMsg {
	return &CasteMsg{cp.PId, cp.CId, msg}
}

func (cp *CasteProcess) StopListen() {
	select {
	case cp.StopChan <- true:
	default:
	}
	cp.UnicastListener.Close()

	select {
	case cp.StopChan <- true:
	default:
	}
	cp.MulticastListener.Close()

	select {
	case cp.StopChan <- true:
	default:
	}
	cp.BroadcastListener.Close()
}

func (cp *CasteProcess) CheckLeader() {
	go func() {
		for cp.Status == "WaitingElection" {
			log.Printf("(P:%d-%d) Waiting election! Will check in %d milliseconds\n", cp.PId, cp.CId, defaultTimeOut)
			time.Sleep(defaultTimeOut * time.Millisecond)
		}
		_, err := unicastSendMessage(cp, cp.Leader, cp.HCId, cp.Msg("AreYouOk?"))
		if err != nil {
			log.Printf("(P:%d-%d) Leader [P:%d-%d] is down!\n", cp.PId, cp.CId, cp.Leader, cp.HCId)
			cp.StartElection(false, cp.CId)
		}
	}()
}

func (cp *CasteProcess) StartElection(byContinue bool, caste int) {
	if cp.CId == cp.HCId {
		cp.BecomeLeader()
	} else {
		if !byContinue {
			multicastSendMessage(cp, caste, cp.Msg("WaitElection!"), false)
		}
		multicastSendMessage(cp, caste+1, cp.Msg("SomeoneUp?"), false)
		cp.Status = "WaitingCandidate"
		cp.mux.Lock()
		cp.Candidate = 0
		cp.mux.Unlock()
		go func() {
			select {
			case <-cp.CandidateChan:
				cp.Status = "WaitingElection"
			case <-time.After(time.Millisecond * defaultTimeOut):
				if caste == cp.HCId-1 {
					log.Printf("(P:%d-%d) No candidates in superior castes", cp.PId, cp.CId)
					cp.BecomeLeader()
				} else {
					cp.StartElection(true, caste+1)
				}
			}
		}()
	}
}

func (cp *CasteProcess) BecomeLeader() {
	log.Printf("(P:%d-%d) I am the new leader\n", cp.PId, cp.CId)
	cp.Leader = cp.PId
	cp.HCId = cp.CId
	cp.Status = "Up"
	cp.StopListen()
	time.Sleep(10 * time.Millisecond)
	cp.Start()
	multicastSendMessage(cp, 0, cp.Msg("IAmLeader!"), true)
}

func (cp *CasteProcess) unicastMsgHandler(n int, b []byte, addr string) []byte {
	var msg CasteMsg
	var returnMsg *CasteMsg
	msg.Decode(string(b[:n]))
	switch msg.Text {
	case "AreYouOk?":
		returnMsg = cp.Msg("ok!")
	case "IAmCandidate!":
		cp.mux.Lock()
		if cp.Candidate == 0 {
			returnMsg = cp.Msg("Continue!")
			cp.Candidate = msg.PId
			cp.mux.Unlock()
			cp.CandidateChan <- msg.PId
		} else {
			returnMsg = cp.Msg("YouLost!")
			cp.mux.Unlock()
		}
	default:
		returnMsg = cp.Msg("ok...")
	}
	log.Printf("(P:%d-%d) Unicast message received from [P:%d-%d] at %s: %s\n", cp.PId, cp.CId, msg.PId, msg.CId, addr, msg.Text)
	return []byte(returnMsg.Encode())
}

func (cp *CasteProcess) multicastMsgHandler(src *net.UDPAddr, n int, b []byte, isBroadcast bool) {
	var msg CasteMsg
	msg.Decode(string(b[:n]))
	if cp.PId != msg.PId {
		if isBroadcast {
			log.Printf("(P:%d-%d) Broadcast message received from [P:%d-%d] at %s: %s\n", cp.PId, cp.CId, msg.PId, msg.CId, src.String(), msg.Text)
		} else {
			log.Printf("(P:%d-%d) Multicast message received from [P:%d-%d] at %s: %s\n", cp.PId, cp.CId, msg.PId, msg.CId, src.String(), msg.Text)
		}
		switch msg.Text {
		case "IAmLeader!":
			cp.Leader = msg.PId
			cp.HCId = msg.CId
			cp.Status = "Up"
		case "WaitElection!":
			cp.Status = "WaitingElection"
		case "SomeoneUp?":
			cp.Status = "WaitingElection"
			response, err := unicastSendMessage(cp, msg.PId, msg.CId, cp.Msg("IAmCandidate!"))
			if err == nil {
				if response.Text == "Continue!" {
					cp.StartElection(true, cp.CId)
				}
			} else {
				log.Println("Erro: ", err)
			}
		default:
		}
	}
}
