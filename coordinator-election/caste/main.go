package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ccfortier/go/coordinator-election/caste/multicast"
	"github.com/ccfortier/go/coordinator-election/caste/unicast"
)

const (
	defaultMulticastAddress = "239.0.0.0:9999"
	defaultUnicastAddress   = ":9000"
)

func handler(w http.ResponseWriter, r *http.Request) {
	webinput := r.URL.Query().Get("cmd")
	switch webinput {
	case "lm":
		fmt.Printf("Listening mcast on %s\n", defaultMulticastAddress)
		go listenMulticast()
	case "sm":
		go sendMulticast(defaultMulticastAddress)
	case "lu":
		fmt.Printf("Listening ucast on %s\n", defaultUnicastAddress)
		go listenUnicast()
	case "su":
		go sendUnicast(defaultUnicastAddress)
	case "stop":
		fmt.Println("bye...")
		os.Exit(0)
	default:
		fmt.Printf("Command not recognized %s!", r.URL.Query().Get("cmd"))
		fmt.Println("")
	}
}

func main() {
	fmt.Println("Waiting commands...")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func listenMulticast() {
	multicast.Listen(defaultMulticastAddress, msgHandlerUDP)
}

func msgHandlerUDP(src *net.UDPAddr, n int, b []byte) {
	fmt.Println(string(b[:n]) + " from " + src.String())
}

func sendMulticast(addr string) {
	conn, err := multicast.NewSender(addr)
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte("sm"))
}

func listenUnicast() {
	unicast.Listen(defaultUnicastAddress, msgHandlerTCP)
}

func msgHandlerTCP(n int, b []byte) {
	fmt.Println(string(b[:n]))
}

func sendUnicast(addr string) {
	conn, err := unicast.NewSender(addr)
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte("su"))
}
