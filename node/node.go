package node

import (
	"federated/encryption"
	"federated/transport"
	"fmt"
)

type Node struct {
	Socket  transport.Socket
	Packets []transport.Packet

	encryption.Client
}

func Create() Node {
	n := Node{
		Packets: make([]transport.Packet, 0),
		Client:  encryption.NewClient(),
	}
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

			n.OnReceive(pkt)

			return nil
		}
	}
}

// Handler of packet
func (n *Node) OnReceive(pkt transport.Packet) error {
	n.Packets = append(n.Packets, pkt)

	// If 2 packets -> Calculations + Send back
	if len(n.Packets) >= 2 {

	}

	return nil
}
