package node

import (
	"federated/encryption"
	"federated/neural"
	"federated/transport"
	"fmt"

	"github.com/ldsec/lattigo/v2/ckks"
)

type Node struct {
	// transport
	Socket  transport.Socket
	Packets []transport.Packet

	// encryption
	encryption.Client
	encryption.Server

	// neural
	neural.NeuralNetwork

	// Node
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

// UNFONCTIONAL, TO BE REPAIRED
func CreateAndStart() (Node, error) {
	n := Create()
	err := n.Start() // surely needs a go ?
	return n, err
}

func (n *Node) Print() {
	fmt.Println("Node address : " + n.Socket.GetAddress())
}

func (n *Node) Start() error {
	defer fmt.Println(n.Socket.GetAddress(), "Started without error")

	// TODO : start on create ?

	// Main goroutine of node -> waits for packets
	go func() {
		defer fmt.Println(n.Socket.GetAddress(), "Stopped")
		for {
			pkt, err := n.Socket.Recv()
			if err != nil {
				continue
			}

			// Not saving ACk packets
			if pkt.Type != transport.Ack {
				// TODO save packets per types ?
				n.OnReceive(pkt)
			}

		}
	}()
	return nil
}

func (n *Node) Join(server string) error {
	pktJoin := transport.Packet{
		Source:      n.Socket.GetAddress(),
		Destination: server,
		Type:        transport.Join,
	}
	err := n.Socket.Send(server, pktJoin)
	if err != nil {
		return err
	}

	return nil
}

// Weights are sent with a separator between them, in a string
func (n *Node) SendWeights(server string) error {
	weights := n.NeuralNetwork.GetWeights()
	encrypted, err := n.Client.Encrypt(weights)
	if err != nil {
		return err
	}

	pkt := transport.Packet{
		Source:      n.Socket.GetAddress(),
		Destination: server,
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}

	// Send to server
	err = n.Socket.Send(server, pkt)
	return err
}

// Handler of packet
func (n *Node) OnReceive(pkt transport.Packet) error {

	// TODO switch
	// For testing purpose
	if pkt.Type == "" {
		n.Packets = append(n.Packets, pkt)
	}
	if pkt.Type == transport.Join {
		n.Server.Participants = append(n.Server.Participants, pkt.Source)
		n.Packets = append(n.Packets, pkt)
	}
	if pkt.Type == transport.EncryptedChunk {
		n.Packets = append(n.Packets, pkt)
	}
	if pkt.Type == transport.Result {
		n.Packets = append(n.Packets, pkt)
		cipher := ckks.NewCiphertext(n.Params, 1, 1, 0.01)
		encryption.UnmarshalFromBase64(cipher, pkt.Message)
		coeffs := n.DecodeCoeffs(n.DecryptNew(cipher))
		n.SetWeights(coeffs)
	}

	// TODO geenralize to n participants
	// If 2 packets -> Calculations + Send back
	if len(n.GetPacketsByType(transport.EncryptedChunk)) >= len(n.Server.Participants) && len(n.Server.Participants) > 0 {
		fmt.Println("Server calculations on", len(n.Server.Participants), "polynomes")
		encryptedPkts := n.GetPacketsByType(transport.EncryptedChunk)
		cipherText1 := new(ckks.Ciphertext)
		cipherText2 := new(ckks.Ciphertext)
		encryption.UnmarshalFromBase64(cipherText1, encryptedPkts[0].Message)
		encryption.UnmarshalFromBase64(cipherText2, encryptedPkts[1].Message)
		n.Server.Responses = append(n.Server.Responses, cipherText1)
		n.Server.Responses = append(n.Server.Responses, cipherText2)

		// Server calculations -> averages the weights
		fmt.Println(n.Server.Responses[0])
		adds := n.Server.AddNew(n.Server.Responses[0], n.Server.Responses[1])
		fmt.Println(adds)
		n.Server.Result = n.Server.MultByConstNew(adds, 0.5)

		// Results
		resultsCipher := encryption.MarshalToBase64String(n.Server.Result)

		// Send // Multicast
		for _, p := range encryptedPkts {
			pktResult := transport.Packet{
				Source:      n.Socket.GetAddress(),
				Destination: p.Source,
				Message:     resultsCipher,
				Type:        transport.Result,
			}

			go n.Socket.Send(p.Source, pktResult)
		}

		// Empty used packets
		encryptedPkts = encryptedPkts[0:]
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
