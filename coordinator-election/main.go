package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ccfortier/go/coordinator-election/caste"
)

const (
	defaultMulticastAddress = "239.0.0.0:9999"
	defaultUnicastAddress   = "0.0.0.0:9000"
)

var (
	pCaste     = caste.CasteProcess{}
	admPort    *int
	stopChanel chan bool
	f          *os.File
)

func casteCreate(r *http.Request) {
	pCaste.PId, _ = strconv.Atoi(r.URL.Query().Get("PId"))
	pCaste.CId, _ = strconv.Atoi(r.URL.Query().Get("CId"))
	pCaste.HCId, _ = strconv.Atoi(r.URL.Query().Get("HCId"))
	pCaste.Coordinator, _ = strconv.Atoi(r.URL.Query().Get("Coordinator"))
	pCaste.SingleIP, _ = strconv.Atoi(r.URL.Query().Get("SingleIP"))
	pCaste.Status = "Up"
	pCaste.StopChan = make(chan bool, 1000)
	pCaste.CandidateChan = make(chan int)
	pCaste.FLog = f
	//log.Printf("(P:%d) Created\n", pCaste.PId)
}

func handler(w http.ResponseWriter, r *http.Request) {
	webinput := r.URL.Query()["cmd"]
	if webinput != nil {
		switch webinput[0] {
		case "caste":
			casteCreate(r)
			pCaste.Dump()
			pCaste.UnicastListener, pCaste.MulticastListener, pCaste.BroadcastListener = pCaste.Start()
		case "casteCreate":
			casteCreate(r)
		case "casteStart":
			pCaste.Start()
		case "casteDump":
			pCaste.Dump()
		case "casteStopListen":
			pCaste.StopListen()
		case "casteCheckCoordinator":
			pCaste.CheckCoordinator()
		case "stop":
			if pCaste.PId > 0 {
				log.Fatalf("(P:%d) Stopped!\n", pCaste.PId)
			}
			log.Fatalf("<C.E.Daemon> Stopped on port %d!\n", *admPort)
		default:
			log.Printf("<C.E.Daemon> Command not recognized %s!\n", r.URL.Query().Get("cmd"))
		}
	}
}

func main() {
	var err error
	f, err = os.OpenFile("msglog", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening file: %v\n", err)
	}
	defer f.Close()

	http.HandleFunc("/", handler)
	admPort = flag.Int("admPort", 8080, "Defines http adm port.")
	flag.Parse()
	log.Printf("<C.E.Daemon> waiting commands on port %d...\n", *admPort)
	http.ListenAndServe(fmt.Sprintf(":%d", *admPort), nil)
}
