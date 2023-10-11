package test

import (
	"bufio"
	"fmt"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type client struct {
	Id   party.ID
	Conn chan<- string
}

//for p2p network

type NetworkP2P struct {
	id           party.ID
	idm          party.IDSlice
	mtx          sync.Mutex
	done         chan struct{}
	cmd          chan string
	CmdBroadcast chan string
	messageIn    chan *protocol.Message
	messagesOut  chan *protocol.Message
	entering     chan client
	leaving      chan client
	clients      map[client]bool
}

func NewNetworkP2P(id party.ID, idm party.IDSlice) *NetworkP2P {
	c := &NetworkP2P{
		id:           id,
		idm:          idm,
		done:         make(chan struct{}),
		entering:     make(chan client),
		leaving:      make(chan client),
		cmd:          make(chan string),
		CmdBroadcast: make(chan string),
		messagesOut:  make(chan *protocol.Message), // all incoming client messages
		messageIn:    make(chan *protocol.Message, 9),
		clients:      make(map[client]bool),
	}
	go c.broadcaster()
	return c
}

//命令行消息

func (n *NetworkP2P) CMD() <-chan string {
	return n.cmd
}

//协议接收新的消息

func (n *NetworkP2P) Next(id party.ID) <-chan *protocol.Message {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.messageIn
}

//协议向外发送

func (n *NetworkP2P) Send(msg *protocol.Message) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.messagesOut <- msg
}

func (n *NetworkP2P) Done(id party.ID) chan struct{} {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.done
}

func (n *NetworkP2P) Quit(id party.ID) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
}

func (n *NetworkP2P) broadcaster() {
	// all connected clients
	for {
		select {
		case msg := <-n.messagesOut:
			//log.Printf("Party %s broad to party count:%d", n.id, len(clients))
			for cli := range n.clients {
				if msg.IsFor(cli.Id) {
					if mj, err := msg.MarshalJson(); err == nil {
						cli.Conn <- string(mj)
					} else {
						log.Printf("Send error:%s", err)
					}
				} else {
					log.Printf("Ignore cli %s msg:%v", cli.Id, msg)
				}
			}

		case msg := <-n.CmdBroadcast:
			for cli := range n.clients {
				cli.Conn <- msg
			}
		case cli := <-n.entering:
			log.Printf("Entering party[%s]....", cli.Id)
			n.clients[cli] = true

		case cli := <-n.leaving:
			log.Printf("Leaving party[%s].....", cli.Id)
			delete(n.clients, cli)
			close(cli.Conn)
		}
	}
}

func (n *NetworkP2P) HandleConn(conn net.Conn, who party.ID) {
	// outgoing client messages
	ch := make(chan string)
	go clientWriter(conn, ch)

	//who := conn.RemoteAddr().String()
	id := n.idm[len(n.clients)]

	cli := client{id, ch}

	n.entering <- cli

	n.processMessage(conn)

	// NOTE: ignoring potential errors from input.Err()
	n.leaving <- cli
	err := conn.Close()
	if err != nil {
		log.Printf("Close client error:%s", err)
	}

}

func (n *NetworkP2P) processMessage(conn net.Conn) {
	//每个链接收到的消息，统一放入messages
	bufReader := bufio.NewReader(conn)
	for {
		line, err := bufReader.ReadString('\n')
		if err == nil || err == io.EOF {
			line = strings.TrimRight(line, "\n")
			if line != "" {
				cmdline := strings.Split(line, ":")
				switch {
				case cmdline[0] == "cmd":
					n.CmdBroadcast <- fmt.Sprintf("b%s", line)
					n.cmd <- line
				case cmdline[0] == "bcmd":
					n.cmd <- line
				default:
					msg := protocol.Message{}
					if msg.UnmarshalJson([]byte(line)) != nil {
						log.Printf("handle error:%s", msg.UnmarshalJson([]byte(line)))
					} else {
						n.messageIn <- &msg
					}
				}
			}
			if err == io.EOF {
				break
			}
		} else {
			fmt.Println(err)
			break
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		c, err := conn.Write([]byte(msg))
		if err != nil {
			log.Println(err)
		} else if c > 0 {
			conn.Write([]byte("\n"))
		}
	}
}
