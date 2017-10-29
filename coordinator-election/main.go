package main

import (
	"log"
	"net"
	"net/http"
	"os"
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
	pCaste = caste.CasteProcess{}
)

func handler(w http.ResponseWriter, r *http.Request) {
	webinput := r.URL.Query()["cmd"]
	ip := r.URL.Query()["ip"]
	if webinput != nil {
		switch webinput[0] {
		case "lm":
			log.Printf("Listening mcast on %s\n", defaultMulticastAddress)
			go listenMulticast()
		case "sm":
			go sendMulticast(defaultMulticastAddress)
		case "lu":
			log.Printf("Listening ucast on %s\n", defaultUnicastAddress)
			go listenUnicast()
		case "su":
			if ip != nil {
				go sendUnicast(ip[0])
			}
		case "caste":
			pCaste.PId, _ = strconv.Atoi(r.URL.Query()["PId"][0])
			//pCaste.CId, _ = strconv.Atoi(r.URL.Query()["CId"][0])
			//pCaste.HCId, _ = strconv.Atoi(r.URL.Query()["HCId"][0])
			pCaste.Coordinator, _ = strconv.Atoi(r.URL.Query()["Coordinator"][0])
			pCaste.Start()
		case "stop":
			log.Println("bye...")
			os.Exit(0)
		default:
			log.Printf("Command not recognized %s!\n", r.URL.Query().Get("cmd"))
		}
	}
}

func main() {
	log.Println("Waiting commands...")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
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

func msgHandlerTCP(n int, b []byte) string {
	log.Println(string(b[:n]))
	return "msg received"
}

func sendUnicast(addr string) {
	conn, err := unicast.NewSender(addr)
	if err != nil {
		log.Printf("%s, when try ucast for %s.\n", err, addr)
	} else {
		conn.Write([]byte("su to " + conn.RemoteAddr().String()))
	}
}
