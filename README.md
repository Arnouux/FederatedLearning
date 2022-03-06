# Privacy Preserving Federated Learning
Using multiparty learning and homomorphic encryption, in Go.

## 1. What does this library do ?
The library aims at creating a network of nodes sharing neural network's parameters (weights) between them without sharing them own dataset. Moreover, fully homomorphic encryption is used to limit the access of the server receiving the weights and not let a white-box attack happen.

## 2. How to use ?
You can get inspiration by reading the tests file `test/federated_test.go`.
In any case, you will need to instanciate nodes and make them connect and send data to a server node.
In your project root directory:
`go get github.com/Arnouux/federated-learning-lib`. After that, the library is importable and usable such as:
```
package main

import (
	fl "github.com/Arnouux/federated-learning-lib"
)

func main() {
	node := fl.Create()
	node.Start()

	server := fl.Create()
	server.Start()

	// ** //

	node.Join(server.Socket.GetAddress())

	// ** //

	server.StartLearning()
}
```

## 3. Todo list
- [x] send HE messages
- [x] fragment packets to fit max UDP size
- [x] server can send results back
- [ ] make sure ACKs correspond to current msg / better udp
- [x] nodes can join server / nb of participants 
- [ ] gradients calculations
- [x] aggregation + local weights update
- [ ] byzantine environnement resistance
- [x] change BFV to CKKS for float operations
- [ ] generalize to n participants
- [x] server setup nn
- [ ] let joiners propose input size, hyperparams,.. / not default
- [ ] refactor
- [x] make project importable as lib
  - [ ] make this current repo usable