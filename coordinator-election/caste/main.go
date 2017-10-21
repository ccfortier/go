package main

import (
	"fmt"
	"os"
	"github.com/dmichael/go-multicast/multicast"
	"net"
	"log"
	"encoding/hex"
	"time"
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
		case "listen":
			fmt.Printf("Listening on %s\n", defaultMulticastAddress)
			go listenMulticast()
		case "ping":
			fmt.Printf("Broadcasting to %s\n", defaultMulticastAddress)
			go pingMulticast(defaultMulticastAddress)
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
	log.Println(n, "bytes read from", src)
	log.Println(hex.Dump(b[:n]))
}

func pingMulticast(addr string) {
	conn, err := multicast.NewBroadcaster(addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn.Write([]byte("hello, world\n"))
		time.Sleep(1 * time.Second)
	}
}