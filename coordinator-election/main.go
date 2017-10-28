package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ccfortier/go/multicast"
	"github.com/ccfortier/go/unicast"
)

const (
	defaultMulticastAddress = "239.0.0.0:9999"
	defaultUnicastAddress   = "0.0.0.0:9000"
)

func handler(w http.ResponseWriter, r *http.Request) {
	webinput := r.URL.Query()["cmd"]
	ip := r.URL.Query()["ip"]
	if webinput != nil {
		switch webinput[0] {
		case "lm":
			fmt.Printf("Listening mcast on %s\n", defaultMulticastAddress)
			go listenMulticast()
		case "sm":
			go sendMulticast(defaultMulticastAddress)
		case "lu":
			fmt.Printf("Listening ucast on %s\n", defaultUnicastAddress)
			go listenUnicast()
		case "su":
			if ip != nil {
				go sendUnicast(ip[0])
			}
		case "stop":
			fmt.Println("bye...")
			os.Exit(0)
		default:
			fmt.Printf("Command not recognized %s!\n", r.URL.Query().Get("cmd"))
		}
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
		log.Printf("%s, when try mcast for %s.\n", err, addr)
	} else {
		conn.Write([]byte("sm"))
	}
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
		log.Printf("%s, when try ucast for %s.\n", err, addr)
	} else {
		conn.Write([]byte("su from " + conn.RemoteAddr().String()))
	}
}
