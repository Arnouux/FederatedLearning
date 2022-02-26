package test

import (
	"federated/encryption"
	"federated/neural"
	"federated/node"
	"federated/transport"
	"testing"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/stretchr/testify/require"
)

// Test Send and Recv functions, test should terminate
func Test_SendRecv(t *testing.T) {
	n1 := node.Create()
	//n1.Print()

	n2 := node.Create()
	//n2.Print()

	pkt := transport.Packet{
		Destination: n2.Socket.GetAddress(),
		Message:     "Hello_Test_SendRecv",
		//Payload:     json.RawMessage(`{"Message":"hello"}`),
	}

	recvd := make(chan int)
	go func() {
		_, _ = n2.Socket.Recv()
		//fmt.Println(string(pkt.Message))
		recvd <- 1
	}()
	go n1.Socket.Send(n2.Socket.GetAddress(), pkt)
	_ = <-recvd
	// Should terminate
}

func Test_HE(t *testing.T) {
	client := encryption.NewClient()

	// Encrypt coeffs
	encrypted1, _ := client.Encrypt(2.0, 3.0)
	encrypted2, _ := client.Encrypt(4.0, 5.0)

	server := encryption.NewServer()

	// Send to server
	cipherText1 := new(ckks.Ciphertext)
	cipherText2 := new(ckks.Ciphertext)
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
		Source:      node2.Socket.GetAddress(),
		Destination: node1.Socket.GetAddress(),
		Message:     "Hello_Test_StartNode",
	}

	node2.Socket.Send(node1.Socket.GetAddress(), pkt)

	time.Sleep(time.Millisecond * 200)

	require.Equal(t, pkt, node1.Packets[0])
}

func Test_SendHE(t *testing.T) {
	node1 := node.Create()
	go node1.Start()

	node2 := node.Create()
	go node2.Start()

	encrypted, _ := node1.Client.Encrypt(3, 4)
	pkt := transport.Packet{
		Source:      node2.Socket.GetAddress(),
		Destination: node1.Socket.GetAddress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}

	err := node2.Socket.Send(node1.Socket.GetAddress(), pkt)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 200)
	require.Equal(t, 1, len(node1.Packets))
	require.Equal(t, len(encrypted), len(node1.Packets[0].Message))
	require.Equal(t, pkt.Destination, node1.Packets[0].Destination)
	require.Equal(t, pkt.Message, node1.Packets[0].Message)
	require.Equal(t, pkt.Source, node1.Packets[0].Source)
	require.Equal(t, pkt.Type, node1.Packets[0].Type)
}

func Test_ServerCalculations(t *testing.T) {
	node1 := node.Create()
	node1.Start()
	node2 := node.Create()
	node2.Start()

	server := node.Create()
	server.Server.AddParticipants(node1.Socket.GetAddress(), node2.Socket.GetAddress())
	server.Start()
	serverEncryption := encryption.NewServer()

	// Node 1 encrypts and sends
	encrypted, _ := node1.Client.Encrypt(3, 4)
	pkt := transport.Packet{
		Source:      node1.Socket.GetAddress(),
		Destination: server.Socket.GetAddress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}
	err := node1.Socket.Send(server.Socket.GetAddress(), pkt)
	require.NoError(t, err)

	// Node 2 encrypts and sends
	encrypted, _ = node2.Client.Encrypt(5, 10)
	pkt = transport.Packet{
		Source:      node2.Socket.GetAddress(),
		Destination: server.Socket.GetAddress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}
	err = node2.Socket.Send(server.Socket.GetAddress(), pkt)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 200)
	require.Equal(t, 2, len(server.Packets))

	// Server reads encryptions
	cipherText1 := new(ckks.Ciphertext)
	cipherText2 := new(ckks.Ciphertext)
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
	server.Server.AddParticipants(node1.Socket.GetAddress(), node2.Socket.GetAddress())

	err := server.Start()
	require.NoError(t, err)

	// Node 1 encrypts and sends
	encrypted, _ := node1.Client.Encrypt(3, 4)
	pkt := transport.Packet{
		Source:      node1.Socket.GetAddress(),
		Destination: server.Socket.GetAddress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}

	node1.Start()
	node2.Start()

	err = node1.Socket.Send(server.Socket.GetAddress(), pkt)
	require.NoError(t, err)

	// Node 2 encrypts and sends
	encrypted2, _ := node2.Client.Encrypt(5, 10)
	pkt2 := transport.Packet{
		Source:      node2.Socket.GetAddress(),
		Destination: server.Socket.GetAddress(),
		Message:     encrypted2,
		Type:        transport.EncryptedChunk,
	}

	err = node2.Socket.Send(server.Socket.GetAddress(), pkt2)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 200)
	require.Equal(t, 2, len(server.Packets))

	// Nodes 1 & 2 should receive Result packet
	require.Equal(t, 1, len(node1.Packets))
	require.Equal(t, transport.Result, node1.Packets[0].Type)
	require.Equal(t, 1, len(node2.Packets))
	require.Equal(t, transport.Result, node2.Packets[0].Type)

	// Node 1 receives
	coeffs, err := node1.Decrypt(node1.Packets[0].Message)
	require.NoError(t, err)

	// Node 2 receives
	coeffs2, err2 := node2.Decrypt(node2.Packets[0].Message)
	require.NoError(t, err2)

	require.Equal(t, coeffs[0], coeffs2[0])
	require.Equal(t, coeffs[1], coeffs2[1])
	require.Equal(t, int64(10+4), coeffs2[1])
	require.Equal(t, int64(5+3), coeffs2[0])
}

func Test_ServerWaitsForNodes(t *testing.T) {
	server := node.Create()
	server.Print()

	// Server starts and wait for nodes to join
	server.Start()

	node1 := node.Create()
	node1.Join(server.Socket.GetAddress())

	time.Sleep(time.Millisecond * 200)

	require.Equal(t, 1, len(server.Packets))
	require.Equal(t, 1, len(server.Server.Participants))

	// TODO more about Join
}

func Test_NeuralNetwork(t *testing.T) {
	// input = 4
	// output = 1
	// nb of hidden layers = 1
	// nb of neurons = 5
	// lr = 0.01
	nn := neural.CreateNetwork(4, 1, 1, 5, 0.01)
	nn.InitiateWeights()

	// Change weights to known value
	nn.Weights[0][0] = []float64{1, 1, 1, 1, 1}
	nn.Weights[0][1] = []float64{1, 1, 1, 1, 1}
	nn.Weights[0][2] = []float64{1, 1, 1, 1, 1}
	nn.Weights[0][3] = []float64{1, 1, 1, 1, 1}
	nn.Weights[2][0] = []float64{1, 1, 1, 1, 1}

	nn.Print()
	// input -> hidden1 / hidden1 -> output
	require.Equal(t, 2, len(nn.Weights))

	output, err := nn.Forward([]float64{0.01, 0.02, 0.03, 0.04})
	require.NoError(t, err)
	require.Equal(t, 0.93244675427215695, output[2][0])

	backpropagation, err := nn.Backpropagation([]float64{0.01, 0.02, 0.03, 0.04}, 1)
	require.NoError(t, err)
	require.Equal(t, backpropagation, [][]float64{
		[]float64{-0.002233873461468998, 0, 0, 0},
		[]float64{-1.0611363867362362e-05, -2.1222727734724723e-05, -3.183409160208709e-05, -4.2445455469449446e-05},
	})
}

func Test_Weights(t *testing.T) {
	n := node.Create()
	n.NeuralNetwork = neural.CreateNetwork(4, 1, 1, 5, 0.01)
	n.NeuralNetwork.InitiateWeights()
	n.NeuralNetwork.Print()

	require.Equal(t, 1, 2)
}

// Prepare number of layers, number of neurons,
// learning rate, number of global iterations,
// activation functions, local batch size.
// Root (server) encrypts initial weigths.
func Test_ServerPreparesParameters(t *testing.T) {
	server := node.Create()
	server.Start()

	node1, err1 := node.CreateAndStart()
	node1.Join(server.Socket.GetAddress())
	node2, err2 := node.CreateAndStart()
	node2.Join(server.Socket.GetAddress())
	require.NoError(t, err1, err2)

	// TODO : See Protocol 1 Collective Training
}

func Test_LocalGradientDescent(t *testing.T) {
	// TODO : See Protocol 2 LGD
}

func Test_UpdateLocalModel(t *testing.T) {
	// After receiving aggregated weights,
	// decrypt and apply to the node's nn.
}

func Test_CombineGradients(t *testing.T) {
	// sum all received gradients

	// Update model weights by using
	// the averaged aggragated gradients
}
