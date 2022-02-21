package neural

import (
	"fmt"
	"math/rand"
)

type NeuralNetwork struct {
	InputDimensions    int
	OutputDimensions   int
	NbLayers           int
	NbNeurons          int
	LearningRate       float64
	NbIterations       int
	ActivationFunction string
	BatchSize          int
	Weights            map[int]map[int][]float64 // layer / index neuron / array of weigths
}

type Neuron struct {
}

const (
	Sigmoid = "sigmoid"
	Relu    = "relu"
)

func CreateNetwork(input int, output int,
	nbLayers int, nbNeurons int, lr float64,
	activationFunction string) NeuralNetwork {

	nn := NeuralNetwork{
		InputDimensions:    input,
		OutputDimensions:   output,
		NbLayers:           nbLayers,
		NbNeurons:          nbNeurons,
		LearningRate:       lr,
		ActivationFunction: activationFunction,
	}

	return nn
}

func (nn *NeuralNetwork) InitiateWeights() {
	// weigths are in [0.0, 1.0)
	nn.Weights = make(map[int]map[int][]float64)

	// Input layer
	nn.Weights[0] = make(map[int][]float64)
	for input := 0; input < nn.InputDimensions; input++ {
		nn.Weights[0][input] = RandomArrayFloat(nn.NbNeurons)
	}

	// Hidden layers
	for layer := 1; layer <= nn.NbLayers; layer++ {
		nn.Weights[layer] = make(map[int][]float64)
		for neuron := 0; neuron < nn.NbNeurons; neuron++ {
			nn.Weights[layer][neuron] = RandomArrayFloat(nn.NbNeurons)
		}
	}

	// Output layer
	nn.Weights[nn.NbLayers+1] = make(map[int][]float64)
	for input := 0; input < nn.OutputDimensions; input++ {
		nn.Weights[nn.NbLayers+1][input] = RandomArrayFloat(nn.NbNeurons)
	}
}

func (nn *NeuralNetwork) Print() {
	for input := 0; input < nn.InputDimensions; input++ {
		fmt.Println(nn.Weights[0][input])
	}
	for layer := 1; layer <= nn.NbLayers; layer++ {
		for neuron := 0; neuron < nn.NbNeurons; neuron++ {
			fmt.Println(nn.Weights[layer][neuron])
		}
	}
	for input := 0; input < nn.OutputDimensions; input++ {
		fmt.Println(nn.Weights[nn.NbLayers+1][input])
	}
}

// UTILS

func RandomArrayFloat(size int) []float64 {
	randomArray := make([]float64, size)
	for i := range randomArray {
		randomArray[i] = rand.Float64()
	}
	return randomArray
}
