package test

import (
	"federated/encryption"
	"federated/neural"
	"federated/node"
	"federated/transport"
	"fmt"
	"testing"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/stretchr/testify/require"
)

// Test Send and Recv functions
func Test_SendRecv(t *testing.T) {
	n1 := node.Create()
	n2 := node.Create()

	pkt := transport.Packet{
		Destination: n2.Socket.GetAddress(),
		Message:     "Hello_Test_SendRecv",
		//Payload:     json.RawMessage(`{"Message":"hello"}`),
	}

	recvd := make(chan int)
	go func() {
		_, _ = n2.Socket.Recv()
		recvd <- 1
	}()
	go n1.Socket.Send(n2.Socket.GetAddress(), pkt)
	_ = <-recvd
	// Should terminate
}

func Test_HE(t *testing.T) {
	client := encryption.NewClient()

	// Encrypt coeffs
	encrypted2, _ := client.Encrypt([]float64{4, 5})

	server := encryption.NewServer()

	// Send to server
	p := ckks.DefaultParams[1]
	params, _ := ckks.NewParametersFromLiteral(p)
	cipherText1 := ckks.NewCiphertext(params, 100, 1, 0.01)
	cipherText2 := new(ckks.Ciphertext)
	encryption.UnmarshalFromBase64(cipherText2, encrypted2)
	server.Responses = append(server.Responses, cipherText1)
	server.Responses = append(server.Responses, cipherText2)

	// Plaintext creation and encoding process
	plaintext := ckks.NewPlaintext(client.Params, client.Params.MaxLevel(), client.Params.DefaultScale())
	client.Encoder.EncodeCoeffs([]float64{1, 2, 3}, plaintext)
	// Encryption process
	ciphertext := client.Encryptor.EncryptNew(plaintext)
	// Calcs
	adds := server.Evaluator.AddNew(ciphertext, ciphertext)
	mean := server.Evaluator.MultByConstNew(adds, 0.5)
	// For transport
	cipher := encryption.MarshalToBase64String(mean)
	// Received
	cipher2 := ckks.NewCiphertext(client.Params, 1, 1, 0.01)
	encryption.UnmarshalFromBase64(cipher2, cipher)
	// Decryption process
	coefs := client.Encoder.DecodeCoeffs(client.Decryptor.DecryptNew(cipher2))
	coefs = encryption.RemoveZerosCoeffs(coefs)
	fmt.Println(int(coefs[0]+0.5), int(coefs[1]+0.5), int(coefs[2]+0.5))
	require.Equal(t, 1, int(coefs[0]+0.5))
	require.Equal(t, 2, int(coefs[1]+0.5))
	require.Equal(t, 3, int(coefs[2]+0.5))

	// ----
	// // Server calculations
	// fmt.Println(server.MulNew(server.Responses[0], server.Responses[1]))
	// server.Responses = append(server.Responses, server.RelinearizeNew(server.MulNew(server.Responses[0], server.Responses[1])))
	// server.Result = server.Responses[2]
	// // Results
	// cipherResult := encryption.MarshalToBase64String(server.Result)
	// coeffs, _ := client.Decrypt(cipherResult)
	// require.Equal(t, int64(2*4), coeffs[0])
	// require.Equal(t, int64(3*5), coeffs[1])
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

	encrypted, _ := node1.Client.Encrypt([]float64{3, 4})
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
	encrypted, _ := node1.Client.Encrypt([]float64{3, 4})
	pkt := transport.Packet{
		Source:      node1.Socket.GetAddress(),
		Destination: server.Socket.GetAddress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}
	err := node1.Socket.Send(server.Socket.GetAddress(), pkt)
	require.NoError(t, err)

	// Node 2 encrypts and sends
	encrypted, _ = node2.Client.Encrypt([]float64{5, 10})
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

	// later : make serverEncryptor in server node
	serverEncryption.Responses = append(serverEncryption.Responses, cipherText1)
	serverEncryption.Responses = append(serverEncryption.Responses, cipherText2)
	// Server calculations
	adds := serverEncryption.AddNew(serverEncryption.Responses[0], serverEncryption.Responses[1])
	serverEncryption.Result = serverEncryption.MultByConstNew(adds, 0.5)

	fmt.Println(serverEncryption.Result)

	// Results
	cipherResult := encryption.MarshalToBase64String(serverEncryption.Result)
	coeffs, _ := node1.Decrypt(cipherResult)

	require.GreaterOrEqual(t, 4+0.000001, coeffs[0])
	require.LessOrEqual(t, 4-0.000001, coeffs[0])
	require.GreaterOrEqual(t, 7+0.000001, coeffs[1])
	require.LessOrEqual(t, 7-0.000001, coeffs[1])
}

func Test_ServerSendResults(t *testing.T) {
	node1 := node.Create()
	node2 := node.Create()

	server := node.Create()
	server.Server.AddParticipants(node1.Socket.GetAddress(), node2.Socket.GetAddress())

	err := server.Start()
	require.NoError(t, err)

	// Node 1 encrypts and sends
	encrypted, _ := node1.Client.Encrypt([]float64{3, 4, 5})
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
	encrypted2, _ := node2.Client.Encrypt([]float64{5, 10, 0})
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
	require.Equal(t, coeffs[2], coeffs2[2])
	require.GreaterOrEqual(t, 4+0.000001, coeffs[0])
	require.LessOrEqual(t, 4-0.000001, coeffs[0])
	require.GreaterOrEqual(t, 7+0.000001, coeffs[1])
	require.LessOrEqual(t, 7-0.000001, coeffs[1])
	require.GreaterOrEqual(t, 2.5+0.000001, coeffs[2])
	require.LessOrEqual(t, 2.5-0.000001, coeffs[2])
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
	n1 := node.Create()
	go n1.Start()
	n2 := node.Create()
	go n2.Start()
	n1.NeuralNetwork = neural.CreateNetwork(4, 1, 1, 5, 0.01)
	n1.NeuralNetwork.InitiateWeights()

	err := n1.SendWeights(n2.Socket.GetAddress(), false)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 200)

	require.Equal(t, 1, len(n2.GetPacketsByType(transport.EncryptedChunk)))

}

func Test_Send2Times(t *testing.T) {
	n1 := node.Create()
	go n1.Start()
	n2 := node.Create()
	go n2.Start()
	err := n1.SendWeights(n2.Socket.GetAddress(), false)
	require.NoError(t, err)
	err = n1.SendWeights(n2.Socket.GetAddress(), false)
	require.NoError(t, err)

	encrypted, _ := n1.Client.Encrypt([]float64{3, 4})
	pkt := transport.Packet{
		Source:      n1.Socket.GetAddress(),
		Destination: n2.Socket.GetAddress(),
		Message:     encrypted,
		Type:        transport.EncryptedChunk,
	}
	n1.Socket.Send(n2.Socket.GetAddress(), pkt)
	n1.Socket.Send(n2.Socket.GetAddress(), pkt)

	time.Sleep(time.Millisecond * 200)
	require.Equal(t, 4, len(n2.Packets))
}

// Prepare number of layers, number of neurons,
// learning rate, number of global iterations,
// activation functions, local batch size.
// Root (server) encrypts initial weigths.
func Test_ServerPreparesParameters(t *testing.T) {
	server := node.Create()
	go server.Start()

	node1 := node.Create()
	go node1.Start()
	node1.Join(server.Socket.GetAddress())
	node2 := node.Create()
	go node2.Start()
	node2.Join(server.Socket.GetAddress())

	// See Protocol 1 Collective Training

	// For now, server directly sends hyperparams
	pktParams := transport.Packet{
		Source:      server.Socket.GetAddress(),
		Destination: node2.Socket.GetAddress(),
		Params: transport.Parameters{
			InputDimensions:    4,
			OutputDimensions:   1,
			NbLayers:           1,
			NbNeurons:          5,
			LearningRate:       0.01,
			NbIterations:       5,
			ActivationFunction: neural.SigmoidFunc,
			BatchSize:          64,
		},
		Type: transport.Params,
	}
	time.Sleep(time.Millisecond * 200)

	require.Equal(t, 1, len(node1.Packets))
	require.Equal(t, pktParams, node2.Packets[0])

	require.Equal(t, node1.Packets[0].Params, node2.Packets[0].Params)
}

func Test_LocalGradientDescent(t *testing.T) {
	// TODO
}

// After receiving aggregated weights,
// decrypt and apply to the node's nn.
func Test_UpdateLocalModel(t *testing.T) {
	n1 := node.Create()
	n1.Start()
	n2 := node.Create()
	n2.Start()
	server := node.Create()
	server.Start()
	n1.Join(server.Socket.GetAddress())
	n2.Join(server.Socket.GetAddress())
	time.Sleep(time.Millisecond * 200)

	n1.NeuralNetwork = neural.CreateNetwork(4, 1, 1, 5, 0.01)
	n2.NeuralNetwork = neural.CreateNetwork(4, 1, 1, 5, 0.01)
	n1.InitiateWeights()
	n2.InitiateWeights()

	// Save current weights
	w1 := n1.GetWeights()
	n1.SendWeights(server.Socket.GetAddress(), false)
	n2.SendWeights(server.Socket.GetAddress(), false)
	time.Sleep(time.Millisecond * 200)
	require.Equal(t, 2, len(server.GetPacketsByType(transport.EncryptedChunk)))
	require.Equal(t, 1, len(n1.GetPacketsByType(transport.Result)))

	// Weights have changed
	require.NotEqual(t, w1, n1.GetWeights())
}

func Test_StartLearning(t *testing.T) {
	n1 := node.Create()
	n1.Start()
	n2 := node.Create()
	n2.Start()
	server := node.Create()
	server.Start()
	n1.Join(server.Socket.GetAddress())
	n2.Join(server.Socket.GetAddress())
	time.Sleep(time.Millisecond * 100)

	require.NotEqual(t, n1.GetWeights(), n2.GetWeights())
	server.StartLearning()

	time.Sleep(time.Millisecond * 100)

	// 2 = params + weights
	require.Equal(t, 2, len(n1.Packets))
	// precision is up to ~10^7
	require.Equal(t, int(n1.GetWeights()[0]*1000), int(n2.GetWeights()[0]*1000))
}
