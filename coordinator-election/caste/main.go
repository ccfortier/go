package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/dmichael/go-multicast/multicast"
)

const (
	defaultMulticastAddress = "239.0.0.0:9999"
)

func main() {
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
	conn, err := multicast.NewBroadcaster(addr)
	if err != nil {
		log.Fatal(err)
	}

	conn.Write([]byte("sm"))

}
