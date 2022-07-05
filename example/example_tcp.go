package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	local := flag.String("local", "0.0.0.0:7000", "local listening addr")
	addr1 := flag.String("addr1", "0.0.0.0:7000", "party1 dial addr")
	addr2 := flag.String("addr2", "0.0.0.0:7000", "party2 dial addr")

	log.Print("List Server ...")
	conn, err := net.Listen("tcp", *local)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Party1 Dial ...")
	party1, err := net.Dial("tcp", *addr1)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Party2 Dial ...")
	party2, err := net.Dial("tcp", *addr2)
	if err != nil {
		log.Fatal(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown Server ...")

}
