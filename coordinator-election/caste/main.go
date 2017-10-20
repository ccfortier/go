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
	defaultMulticastAddress = "224.0.0.100:9999"
)


func main() {
	var input string
	for {
		time.Sleep(10)
		fmt.Print("Enter text: ")
		fmt.Scanln(&input)

		switch input {
		case "listen":
			go listenMulticast()
		case "ping":
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
	fmt.Printf("Listening on %s\n", defaultMulticastAddress)
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