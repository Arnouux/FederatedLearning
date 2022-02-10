package test

import (
	"federated/encryption"
	"federated/node"
	"federated/transport"
	"testing"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/stretchr/testify/require"
)

// Test Send and Recv functions, test should terminate
func Test_SendRecv(t *testing.T) {
	n1 := node.Create()
	//n1.Print()

	n2 := node.Create()
	//n2.Print()

	pkt := transport.Packet{
		Destination: n2.Socket.GetAdress(),
		Message:     "Hello",
		//Payload:     json.RawMessage(`{"Message":"hello"}`),
	}

	recvd := make(chan int)
	go func() {
		_, _ = n2.Socket.Recv()
		//fmt.Println(string(pkt.Message))
		recvd <- 1
	}()
	go n1.Socket.Send(n2.Socket.GetAdress(), pkt)

	_ = <-recvd
	// Should terminate
}

func Test_HE(t *testing.T) {
	client := encryption.NewClient()

	// Encrypt coeffs
	encrypted1, _ := client.Encrypt(2, 3)
	encrypted2, _ := client.Encrypt(4, 5)

	server := encryption.NewServer()

	// Send to server
	cipherText1 := new(bfv.Ciphertext)
	cipherText2 := new(bfv.Ciphertext)
	encryption.UnmarshalFromBase64(cipherText1, encrypted1)
	encryption.UnmarshalFromBase64(cipherText2, encrypted2)
	server.Responses = append(server.Responses, cipherText1)
	server.Responses = append(server.Responses, cipherText2)
	// Server calculations
	server.Responses = append(server.Responses, server.RelinearizeNew(server.MulNew(server.Responses[0], server.Responses[1])))
	server.Result = server.Responses[2]

	// Results
	cipherResult := encryption.MarshalToBase64String(server.Result)
	coeffs, _ := client.Decrypt(cipherResult)

	require.Equal(t, int64(2*4), coeffs[0])
	require.Equal(t, int64(3*5), coeffs[1])
}

func Test_StartNode(t *testing.T) {
	node1 := node.Create()
	go node1.Start()

	node2 := node.Create()

	pkt := transport.Packet{
		Source:      node2.Socket.GetAdress(),
		Destination: node1.Socket.GetAdress(),
		Message:     "Hello",
		//Payload:     json.RawMessage(`{"Message":"hello"}`),
	}

	node2.Socket.Send(node1.Socket.GetAdress(), pkt)

	require.Equal(t, 1, 2)
}
