package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/ccfortier/go/coordinator-election/caste"
	"github.com/ccfortier/go/multicast"
	"github.com/ccfortier/go/unicast"
)

const (
	defaultMulticastAddress = "239.0.0.0:9999"
	defaultUnicastAddress   = "0.0.0.0:9000"
)

var (
	pCaste  = caste.CasteProcess{}
	admPort *int
)

func handler(w http.ResponseWriter, r *http.Request) {
	webinput := r.URL.Query()["cmd"]
	ip := r.URL.Query()["ip"]
	if webinput != nil {
		switch webinput[0] {
		case "lm":
			log.Printf("<C.E.Daemon> listening mcast on %s\n", defaultMulticastAddress)
			go listenMulticast()
		case "sm":
			go sendMulticast(defaultMulticastAddress)
		case "lu":
			log.Printf("<C.E.Daemon> listening ucast on %s\n", defaultUnicastAddress)
			go listenUnicast()
		case "su":
			if ip != nil {
				go sendUnicast(ip[0])
			}
		case "caste":
			pCaste.PId, _ = strconv.Atoi(r.URL.Query().Get("PId"))
			pCaste.CId, _ = strconv.Atoi(r.URL.Query().Get("CId"))
			pCaste.HCId, _ = strconv.Atoi(r.URL.Query().Get("HCId"))
			pCaste.Coordinator, _ = strconv.Atoi(r.URL.Query().Get("Coordinator"))
			pCaste.SingleIP, _ = strconv.Atoi(r.URL.Query().Get("SingleIP"))
			pCaste.Start()
		case "casteDump":
			pCaste.Dump()
		case "casteCheckCoordinator":
			pCaste.CheckCoordinator()
		case "stop":
			log.Fatalf("<C.E.Daemon> stopped on port %d...\n", *admPort)
		default:
			log.Printf("<C.E.Daemon> command not recognized %s!\n", r.URL.Query().Get("cmd"))
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	admPort = flag.Int("admPort", 8080, "Defines http adm port.")
	flag.Parse()
	log.Printf("<C.E.Daemon> waiting commands on port %d...\n", *admPort)
	http.ListenAndServe(fmt.Sprintf(":%d", *admPort), nil)
}

func listenMulticast() {
	multicast.Listen(defaultMulticastAddress, msgHandlerUDP)
}

func msgHandlerUDP(src *net.UDPAddr, n int, b []byte) {
	log.Println(string(b[:n]) + " from " + src.String())
}

func sendMulticast(addr string) {
	conn, err := multicast.NewSender(addr)
	if err != nil {
		log.Printf("%s, when try mcast for %s.\n", err, addr)
	} else {
		conn.Write([]byte("sm"))
	}
}

func listenUnicast() {
	unicast.Listen(defaultUnicastAddress, msgHandlerTCP)
}

func msgHandlerTCP(n int, b []byte, addr string) []byte {
	log.Printf("Message received from %s: %s\n", addr, string(b[:n]))
	return []byte("msg received")
}

func sendUnicast(addr string) {
	conn, err := unicast.NewSender(addr)
	if err != nil {
		log.Printf("%s, when try ucast for %s.\n", err, addr)
	} else {
		conn.Write([]byte("su to " + conn.RemoteAddr().String()))
	}
}
