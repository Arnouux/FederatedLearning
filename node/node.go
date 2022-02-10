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
