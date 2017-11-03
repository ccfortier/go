package caste

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ccfortier/go/multicast"
	"github.com/ccfortier/go/unicast"
)

func processAddress(SingleIP int, PId int) string {
	if SingleIP == 0 {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, PId, defaultUnicastPort)
	} else {
		return fmt.Sprintf("%s.%d:%d", defaultNetwork, SingleIP, defaultUnicastPort+PId)
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
	l, err := unicast.NewListener(processAddress(cp.SingleIP, cp.PId))
	cp.UnicastListener = l
	if err == nil {
		go unicast.Listen(cp.UnicastListener, cp.unicastMsgHandler, cp.StopChan)
	}
}

func unicastSendMessage(cp *CasteProcess, tPID int, tCID int, msg *CasteMsg) (*CasteMsg, error) {
	var response CasteMsg

	log.SetOutput(cp.FLog)
	log.Printf("(P:%d-%d) Unicast: %s to [P:%d-%d]\n", cp.PId, cp.CId, msg.Text, tPID, tCID)
	if *cp.QuietMode {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	conn, err := unicast.NewSender(processAddress(cp.SingleIP, tPID))
	if err != nil {
		return nil, err
	} else {
		conn.Write([]byte(msg.Encode()))
		buffer := make([]byte, maxDatagramSize)
		n, err := conn.Read(buffer)
		if err != nil {
			return nil, fmt.Errorf("(P:%d-%d) ReadFromTCP failed to colect response: %s", cp.PId, cp.CId, err)
		}
		response.Decode(string(buffer[:n]))
		log.Printf("(P:%d-%d) Response from process [P:%d-%d]: %s", cp.PId, cp.CId, response.PId, response.CId, response.Text)
		return &response, nil
	}
}

func multicastListen(cp *CasteProcess) {
	l, err := multicast.NewListener(casteAddress(cp.CId))
	cp.MulticastListener = l
	if err == nil {
		go multicast.Listen(cp.MulticastListener, cp.multicastMsgHandler, false, cp.StopChan)
	}
}

func broadcastListen(cp *CasteProcess) {
	l, err := multicast.NewListener(broadcastAddress())
	cp.BroadcastListener = l
	if err == nil {
		go multicast.Listen(cp.BroadcastListener, cp.multicastMsgHandler, true, cp.StopChan)
	}
}

func multicastSendMessage(cp *CasteProcess, tPID int, msg *CasteMsg, isBroadcast bool) error {
	var addr string
	log.SetOutput(cp.FLog)
	if isBroadcast {
		addr = broadcastAddress()
		log.Printf("(P:%d-%d) Broadcast: %s", cp.PId, cp.CId, msg.Text)
	} else {
		addr = casteAddress(tPID)
		log.Printf("(P:%d-%d) Multicast: %s to caste %d", cp.PId, cp.CId, msg.Text, tPID)
	}
	if *cp.QuietMode {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	conn, err := multicast.NewSender(addr)
	if err != nil {
		return err
	} else {
		conn.Write([]byte(msg.Encode()))
		return nil
	}
}
