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