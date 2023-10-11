package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type client chan<- string // an outgoing message channel

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) // all incoming client messages
)

func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	for {
		select {
		case msg := <-messages:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			log.Printf("broadcast %d client msg:%s", len(clients), msg)
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

//!-broadcaster

//!+handleConn
func handleConn(conn net.Conn) {
	ch := make(chan string) // outgoing client messages
	go clientWriter(conn, ch)

	entering <- ch

	//input := bufio.NewScanner(conn)
	//for input.Scan() {
	//	messages <- input.Text()
	//}
	//if input.Err() != nil {
	//	log.Println("Scan error", input.Err())
	//}

	bufReader := bufio.NewReader(conn)
	for {
		line, err := bufReader.ReadString('\n')
		if err == nil || err == io.EOF {
			if line != "" {
				messages <- line
			}
			if err == io.EOF {
				break
			}
		} else {
			fmt.Println(err)
			break
		}
	}

	// NOTE: ignoring potential errors from input.Err()
	log.Println("Client close.....")
	leaving <- ch
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // NOTE: ignoring network errors
	}
}

//!-handleConn

//!+main
func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

//!-main
