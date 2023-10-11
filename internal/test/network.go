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

type INetwork interface {
	Next(id party.ID) <-chan *protocol.Message
	Send(msg *protocol.Message)
	Done(id party.ID) chan struct{}
	Quit(id party.ID)
}

//for broadcast tcp network

type NetworkBroadCast struct {
	id               party.ID
	conn             net.Conn
	cmdChannel       chan string
	listenChannel    chan *protocol.Message
	writeChannel     chan *protocol.Message
	done             chan struct{}
	closedListenChan chan *protocol.Message
	mtx              sync.Mutex
}

func NewNetworkTcp(id party.ID, conn net.Conn) *NetworkBroadCast {
	closed := make(chan *protocol.Message)
	close(closed)
	c := &NetworkBroadCast{
		id:               id,
		conn:             conn,
		closedListenChan: closed}
	c.init()
	return c
}

func (n *NetworkBroadCast) init() {
	n.done = make(chan struct{})
	n.cmdChannel = make(chan string)
	n.listenChannel = make(chan *protocol.Message, 9)
	n.writeChannel = make(chan *protocol.Message, 9)
	go n.reader()
	go n.clientWriter()
}

func (n *NetworkBroadCast) reader() {
	bufReader := bufio.NewReader(n.conn)
	for {
		line, err := bufReader.ReadString('\n')
		if err == nil || err == io.EOF {
			line = strings.TrimRight(line, "\n")
			n.processLine(line)
			if err == io.EOF {
				break
			}
		} else {
			fmt.Println(err)
			break
		}
	}
}

func (n *NetworkBroadCast) processLine(line string) {
	if line != "" {
		cmdline := strings.Split(line, ":")
		if cmdline[0] == "cmd" {
			n.cmdChannel <- cmdline[1]
		} else {
			n.processMsg(line)
		}
	}
}

func (n *NetworkBroadCast) processMsg(line string) {
	msg := protocol.Message{}
	err := msg.UnmarshalJson([]byte(line))
	if err != nil {
		log.Println(err)
	} else {
		if msg.IsFor(n.id) {
			n.listenChannel <- &msg
		}
	}
}

func (n *NetworkBroadCast) clientWriter() {
	for msg := range n.writeChannel {
		mj, err := msg.MarshalJson()
		if err != nil {
			log.Println(err)
			continue
		}
		count, err := n.conn.Write(mj)
		if err != nil {
			log.Println(err)
		} else if count > 0 {
			_, err = n.conn.Write([]byte("\n"))
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (n *NetworkBroadCast) CMD() <-chan string {
	return n.cmdChannel
}

func (n *NetworkBroadCast) Next(id party.ID) <-chan *protocol.Message {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.listenChannel
}

func (n *NetworkBroadCast) Send(msg *protocol.Message) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.writeChannel <- msg
}

func (n *NetworkBroadCast) Done(id party.ID) chan struct{} {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.done
}

func (n *NetworkBroadCast) Quit(id party.ID) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
}

// Network simulates a point-to-point network between different parties using Go channels.
// The same network is used by all processes, and can be reused for different protocols.
// When used with test.Handler, no interaction from the user is required beyond creating the network.
type Network struct {
	parties          party.IDSlice
	listenChannels   map[party.ID]chan *protocol.Message
	done             chan struct{}
	closedListenChan chan *protocol.Message
	mtx              sync.Mutex
}

func NewNetwork(parties party.IDSlice) *Network {
	closed := make(chan *protocol.Message)
	close(closed)
	c := &Network{
		parties:          parties,
		listenChannels:   make(map[party.ID]chan *protocol.Message, 2*len(parties)),
		closedListenChan: closed,
	}
	return c
}

func (n *Network) init() {
	N := len(n.parties)
	for _, id := range n.parties {
		n.listenChannels[id] = make(chan *protocol.Message, N*N)
	}
	n.done = make(chan struct{})
}

func (n *Network) Next(id party.ID) <-chan *protocol.Message {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	if len(n.listenChannels) == 0 {
		n.init()
	}
	c, ok := n.listenChannels[id]
	if !ok {
		return n.closedListenChan
	}
	return c
}

func (n *Network) Send(msg *protocol.Message) {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	//根据消息ID，循环找到ID对应的通道，发出去
	for id, c := range n.listenChannels {
		if msg.IsFor(id) && c != nil {
			//send to other
			n.listenChannels[id] <- msg
		}
	}
}

func (n *Network) Done(id party.ID) chan struct{} {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	if _, ok := n.listenChannels[id]; ok {
		close(n.listenChannels[id])
		delete(n.listenChannels, id)
	}
	if len(n.listenChannels) == 0 {
		close(n.done)
	}
	return n.done
}

func (n *Network) Quit(id party.ID) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.parties = n.parties.Remove(id)
}
