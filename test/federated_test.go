package test

import (
	"federated/encryption"
	"federated/node"
	"federated/transport"
	"testing"
	"time"

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
		Message:     "Hello_Test_SendRecv",
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
	go node2.Start()

	pkt := transport.Packet{
		Source:      node2.Socket.GetAdress(),
		Destination: node1.Socket.GetAdress(),
		Message:     "Hello_Test_StartNode",
	}

	node2.Socket.Send(node1.Socket.GetAdress(), pkt)

	time.Sleep(time.Second * 1)

	require.Equal(t, pkt, node1.Packets[0])
}

func Test_SendHE(t *testing.T) {
	node1 := node.Create()
	go node1.Start()

	node2 := node.Create()
	go node2.Start()

	encrypted, _ := node1.Client.Encrypt(3, 4)
	pkt := transport.Packet{
		Source:      node2.Socket.GetAdress(),
		Destination: node1.Socket.GetAdress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}

	err := node2.Socket.Send(node1.Socket.GetAdress(), pkt)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)
	require.Equal(t, 1, len(node1.Packets))
	require.Equal(t, len(encrypted), len(node1.Packets[0].Message))
	require.Equal(t, pkt.Destination, node1.Packets[0].Destination)
	require.Equal(t, pkt.Message, node1.Packets[0].Message)
	require.Equal(t, pkt.Source, node1.Packets[0].Source)
	require.Equal(t, pkt.Type, node1.Packets[0].Type)
}

func Test_ServerCalculations(t *testing.T) {
	node1 := node.Create()
	go node1.Start()
	node2 := node.Create()
	go node2.Start()

	server := node.Create()
	go server.Start()
	serverEncryption := encryption.NewServer()

	// Node 1 encrypts and sends
	encrypted, _ := node1.Client.Encrypt(3, 4)
	pkt := transport.Packet{
		Source:      node1.Socket.GetAdress(),
		Destination: server.Socket.GetAdress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}
	err := node1.Socket.Send(server.Socket.GetAdress(), pkt)
	require.NoError(t, err)

	// Node 2 encrypts and sends
	encrypted, _ = node2.Client.Encrypt(5, 10)
	pkt = transport.Packet{
		Source:      node2.Socket.GetAdress(),
		Destination: server.Socket.GetAdress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}
	err = node2.Socket.Send(server.Socket.GetAdress(), pkt)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)
	require.Equal(t, 2, len(server.Packets))

	// Server reads encryptions
	cipherText1 := new(bfv.Ciphertext)
	cipherText2 := new(bfv.Ciphertext)
	encryption.UnmarshalFromBase64(cipherText1, server.Packets[0].Message)
	encryption.UnmarshalFromBase64(cipherText2, server.Packets[1].Message)

	// TODO : make serverEncryptor in server node ?
	serverEncryption.Responses = append(serverEncryption.Responses, cipherText1)
	serverEncryption.Responses = append(serverEncryption.Responses, cipherText2)
	// Server calculations
	serverEncryption.Responses = append(serverEncryption.Responses, serverEncryption.RelinearizeNew(serverEncryption.MulNew(serverEncryption.Responses[0], serverEncryption.Responses[1])))
	serverEncryption.Result = serverEncryption.Responses[2]

	// Results
	cipherResult := encryption.MarshalToBase64String(serverEncryption.Result)
	coeffs, _ := node1.Decrypt(cipherResult)

	require.Equal(t, int64(3*5), coeffs[0])
	require.Equal(t, int64(4*10), coeffs[1])
}

func Test_ServerSendResults(t *testing.T) {
	node1 := node.Create()
	node2 := node.Create()

	server := node.Create()
	err := server.Start()
	require.NoError(t, err)

	// Node 1 encrypts and sends
	encrypted, _ := node1.Client.Encrypt(3, 4)
	pkt := transport.Packet{
		Source:      node1.Socket.GetAdress(),
		Destination: server.Socket.GetAdress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}

	node1.Start()
	node2.Start()

	err = node1.Socket.Send(server.Socket.GetAdress(), pkt)
	require.NoError(t, err)

	// Node 2 encrypts and sends
	encrypted2, _ := node2.Client.Encrypt(5, 10)
	pkt2 := transport.Packet{
		Source:      node2.Socket.GetAdress(),
		Destination: server.Socket.GetAdress(),
		Message:     encrypted2,
		Type:        transport.EncryptedChunk,
	}

	err = node2.Socket.Send(server.Socket.GetAdress(), pkt2)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)
	require.Equal(t, 2, len(server.Packets))

	// Nodes 1 & 2 should receive Result packet
	require.Equal(t, 1, len(node1.Packets))
	require.Equal(t, transport.Result, node1.Packets[0].Type)
	require.Equal(t, 1, len(node2.Packets))
	require.Equal(t, transport.Result, node2.Packets[0].Type)
}
