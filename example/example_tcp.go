package main

//
//import (
//	"bufio"
//	"flag"
//	"fmt"
//	"github.com/taurusgroup/multi-party-sig/internal/test"
//	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
//	"github.com/taurusgroup/multi-party-sig/pkg/party"
//	"github.com/taurusgroup/multi-party-sig/pkg/pool"
//	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
//	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
//	"log"
//	"net"
//	"os"
//	"os/signal"
//	"syscall"
//	"time"
//)
//
//type client chan<- *protocol.Message // an outgoing message channel
//
//var (
//	entering = make(chan client)
//	leaving  = make(chan client)
//	messages = make(chan *protocol.Message) // all incoming client messages
//)
//
//func broadcaster() {
//	clients := make(map[client]bool) // all connected clients
//	for {
//		select {
//		case msg := <-messages:
//			// Broadcast incoming message to all
//			// clients' outgoing message channels.
//
//			for cli := range clients {
//				cli <- msg
//			}
//
//		case cli := <-entering:
//			clients[cli] = true
//
//		case cli := <-leaving:
//			delete(clients, cli)
//			close(cli)
//		}
//	}
//}
//
//func handleConn(conn net.Conn) {
//	ch := make(chan *protocol.Message) // outgoing client messages
//	go clientWriter(conn, ch)
//	entering <- ch
//
//	msg := protocol.Message{}
//	//who := conn.RemoteAddr().String()
//	input := bufio.NewScanner(conn)
//	for input.Scan() {
//		err := msg.UnmarshalBinary(input.Bytes())
//		if err != nil {
//			log.Print(err)
//		}
//		messages <- &msg
//	}
//
//	leaving <- ch
//	//messages <- who + " has left"
//	conn.Close()
//}
//
////write to other clinet
//func clientWriter(conn net.Conn, ch <-chan *protocol.Message) {
//	for msg := range ch {
//		mb, err := msg.MarshalBinary()
//		if err != nil {
//			log.Print(err)
//		}
//		fmt.Fprintln(conn, mb) // NOTE: ignoring network errors
//	}
//}
//
//func CMPKeygen1(id party.ID, ids party.IDSlice, threshold int, n *test.Network, pl *pool.Pool) (*cmp.Config, error) {
//	h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, id, ids, threshold, pl), nil)
//	if err != nil {
//		return nil, err
//	}
//	test.HandlerLoop(id, h, n)
//	r, err := h.Result()
//	if err != nil {
//		return nil, err
//	}
//
//	return r.(*cmp.Config), nil
//}
//
//func main() {
//	local := flag.String("local", "0.0.0.0:7000", "local listening addr")
//	addr1 := flag.String("addr1", "0.0.0.0:7000", "party1 dial addr")
//	addr2 := flag.String("addr2", "0.0.0.0:7000", "party2 dial addr")
//
//	ids := party.IDSlice{"a", "b", "c", "d", "e", "f"}
//	threshold := 4
//	//messageToSign := []byte("hello")
//
//	log.Print("List Server ...")
//	conn, err := net.Listen("tcp", *local)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	go broadcaster()
//
//	go func() {
//		for {
//			conn, err := conn.Accept()
//			if err != nil {
//				log.Print(err)
//				continue
//			}
//			go handleConn(conn)
//		}
//	}()
//
//	net := test.NewNetwork(ids)
//	id := party.ID("a")
//
//	pl := pool.NewPool(0)
//	defer pl.TearDown()
//
//	if cfg, err := CMPKeygen1(id, ids, threshold, net, pl); err != nil {
//		fmt.Println(err)
//	} else {
//		log.Print(cfg.MarshalBinary())
//	}
//
//	time.AfterFunc(time.Minute, func() {
//		//trigger connect to other party
//		log.Print("Party1 Dial ...")
//		_, err = net.Dial("tcp", *addr1)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		log.Print("Party2 Dial ...")
//		_, err = net.Dial("tcp", *addr2)
//		if err != nil {
//			log.Fatal(err)
//		}
//	})
//
//	quit := make(chan os.Signal)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//	log.Print("Shutdown Server ...")
//}
