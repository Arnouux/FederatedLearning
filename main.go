package main

import (
	"fmt"

	"federated/encryption"
	"federated/node"
	"federated/transport"

	"github.com/ldsec/lattigo/v2/bfv"
)

func main() {

	n1 := node.Create()
	n1.Print()

	n2 := node.Create()
	n2.Print()

	pkt := transport.Packet{
		Destination: n2.Socket.GetAdress(),
		Message:     "Hello",
		//Payload:     json.RawMessage(`{"Message":"hello"}`),
	}

	recvd := make(chan int)
	go func() {
		pkt, _ := n2.Socket.Recv()
		fmt.Println(string(pkt.Message))
		recvd <- 1
	}()
	go n1.Socket.Send(n2.Socket.GetAdress(), pkt)

	_ = <-recvd

	client := encryption.NewClient()

	// Encrypt coeffs
	encrypted1, _ := client.Encrypt(2, 3)
	encrypted2, _ := client.Encrypt(4, 5)

	// TODO computation on encrypted inputs, before decrypting the result
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
	fmt.Println(coeffs)
	fmt.Println("end")
}
