package node

import (
	"federated/encryption"
	"federated/transport"
	"fmt"

	"github.com/ldsec/lattigo/v2/bfv"
)

type Node struct {
	Socket  transport.Socket
	Packets []transport.Packet

	encryption.Client
	encryption.Server

	StopChan chan bool
}

func Create() Node {
	n := Node{
		Packets: make([]transport.Packet, 0),
		Client:  encryption.NewClient(),
		Server:  encryption.NewServer(),
	}
	s, err := transport.CreateSocket()
	if err != nil {
		fmt.Println(err)
	}

	n.Socket = s
	return n
}

func CreateAndStart() (Node, error) {
	n := Create()
	err := n.Start()
	return n, err
}

func (n *Node) Print() {
	fmt.Println("Node address : " + n.Socket.GetAdress())
}

func (n *Node) Start() error {
	defer fmt.Println(n.Socket.GetAdress(), "Started without error")

	// TODO : start on create ?

	// Main goroutine of node -> waits for packets
	go func() {
		defer fmt.Println(n.Socket.GetAdress(), "Stopped")
		for {
			pkt, err := n.Socket.Recv()
			if err != nil {
				continue
			}

			// Not saving ACk packets
			if pkt.Type != transport.Ack {
				// TODO save packets per types ?
				n.Packets = append(n.Packets, pkt)

				n.OnReceive(pkt)
			}

		}
	}()
	return nil
}

func (n *Node) Join(server string) error {
	pktJoin := transport.Packet{
		Source:      n.Socket.GetAdress(),
		Destination: server,
		Type:        transport.Join,
	}
	err := n.Socket.Send(server, pktJoin)
	if err != nil {
		return err
	}

	return nil
}

// Handler of packet
func (n *Node) OnReceive(pkt transport.Packet) error {
	fmt.Println("Pkt received: ", pkt.Type, "from", pkt.Source)
	// switch pkt.Type {
	// case transport.EncryptedChunk:
	// 	OnReceiveEncryptedChunk(pkt)
	// }

	// If 2 packets -> Calculations + Send back
	if len(n.GetPacketsByType(transport.EncryptedChunk)) >= len(n.Server.Participants) && len(n.Server.Participants) > 0 {
		fmt.Println("Server calculations on", len(n.Server.Participants), "polynomes")

		cipherText1 := new(bfv.Ciphertext)
		cipherText2 := new(bfv.Ciphertext)
		encryption.UnmarshalFromBase64(cipherText1, n.Packets[0].Message)
		encryption.UnmarshalFromBase64(cipherText2, n.Packets[1].Message)
		n.Server.Responses = append(n.Server.Responses, cipherText1)
		n.Server.Responses = append(n.Server.Responses, cipherText2)

		// Server calculations
		n.Server.Responses = append(n.Server.Responses, n.Server.RelinearizeNew(n.Server.MulNew(n.Server.Responses[0], n.Server.Responses[1])))
		n.Server.Result = n.Server.Responses[2]

		// Results
		resultsCipher := encryption.MarshalToBase64String(n.Server.Result)

		// Send // Multicast
		for _, p := range n.Packets {
			pktResult := transport.Packet{
				Source:      n.Socket.GetAdress(),
				Destination: p.Source,
				Message:     resultsCipher,
				Type:        transport.Result,
			}

			go n.Socket.Send(p.Source, pktResult)

		}
		// Empty used packets ?
	}

	if pkt.Type == transport.Join {
		n.Server.Participants = append(n.Server.Participants, pkt.Source)
	}

	return nil
}

func (n *Node) GetPacketsByType(t string) []transport.Packet {
	pkts := make([]transport.Packet, 0)
	for _, p := range n.Packets {
		if p.Type == t {
			pkts = append(pkts, p)
		}
	}
	return pkts
}
