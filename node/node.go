package node

import (
	"federated/transport"
	"fmt"
)

type Node struct {
	Socket transport.Socket
}

func Create() Node {
	n := Node{}
	s, err := transport.CreateSocket()
	if err != nil {
		fmt.Println(err)
	}
	n.Socket = s
	return n
}

func (n *Node) Print() {
	fmt.Println("Node address : " + n.Socket.GetAdress())
}

func (n *Node) Start() error {
	errChan := make(chan error)

	// Main goroutine of node -> waits for packets
	for {
		pktChan := make(chan transport.Packet)
		go func() {
			pkt, err := n.Socket.Recv()
			if err != nil {
				errChan <- err
			}
			pktChan <- pkt
		}()
		select {
		case err := <-errChan:
			return err
		case pkt := <-pktChan:
			fmt.Println("Pkt received: " + pkt.String())

			// TODO handler for the packet

			return nil
		}
	}
}
