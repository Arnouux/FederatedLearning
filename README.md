## Privacy Preserving Federated Learning
Using multiparty learning and homomorphic encryption, in Go.

# 1. What does this library do ?
The library aims at creating a network of nodes sharing neural network's parameters (weights) between them without sharing them own dataset. Moreover, fully homomorphic encryption is used to limit the access of the server receiving the weights and not let a white-box attack happen.

# 2. How to use ?
You can get inspiration by reading the tests file `test/federated_test.go`.
In any case, you will need to instanciate nodes with `node.CreateAndStart()` and make them connect and send data to a server node.

# 3. Todo list
- [x] send HE messages
- [x] fragment packets to fit max UDP size
- [x] server can send results back
- [ ] make sure ACKs correspond to current msg
- [x] nodes can join server / nb of participants 
- [ ] gradients calculations
- [ ] aggregation
- [ ] byzantine environnement resistance