package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ccfortier/go/coordinator-election/caste/multicast"
)

const (
	defaultMulticastAddress = "239.0.0.0:9999"
)

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Query().Get("cmd"))
	webinput := r.URL.Query().Get("cmd")
	switch webinput {
	case "lm":
		fmt.Printf("Listening mcast on %s\n", defaultMulticastAddress)
		go listenMulticast()
	case "sm":
		go sendMulticast(defaultMulticastAddress)
	case "bye":
		fmt.Println("bye...")
		os.Exit(0)
	default:
		fmt.Fprintf(w, "Commanda not recognized %s!", r.URL.Query().Get("cmd"))
	}
}

func main() {

	http.HandleFunc("/", handler)
	go http.ListenAndServe(":8080", nil)

	var input string
	for {
		fmt.Print("Enter text: ")
		fmt.Scanln(&input)

		switch input {
		case "lm":
			fmt.Printf("Listening mcast on %s\n", defaultMulticastAddress)
			go listenMulticast()
		case "sm":
			go sendMulticast(defaultMulticastAddress)
		case "bye":
			fmt.Println("bye...")
			os.Exit(0)
		default:
			fmt.Println(input)
		}
	}

}

func listenMulticast() {
	multicast.Listen(defaultMulticastAddress, msgHandler)
}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	fmt.Println(string(b[:n]) + " from " + src.String())
}

func sendMulticast(addr string) {
	conn, err := multicast.NewSender(addr)
	if err != nil {
		log.Fatal(err)
	}

	conn.Write([]byte("sm"))

}
