package neural

import (
	"errors"
	"fmt"
	"math"
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

const (
	SigmoidFunc = "sigmoid"
	ReluFunc    = "relu"
)

func CreateNetwork(input int, output int,
	nbLayers int, nbNeurons int, lr float64) NeuralNetwork {

	nn := NeuralNetwork{
		InputDimensions:  input,
		OutputDimensions: output,
		NbLayers:         nbLayers,
		NbNeurons:        nbNeurons,
		LearningRate:     lr,
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
	for layer := 1; layer < nn.NbLayers; layer++ {
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
	for layer := 1; layer < nn.NbLayers; layer++ {
		fmt.Println("-----")
		for neuron := 0; neuron < nn.NbNeurons; neuron++ {
			fmt.Println(nn.Weights[layer][neuron])
		}
	}
	fmt.Println("-----")
	for input := 0; input < nn.OutputDimensions; input++ {
		fmt.Println(nn.Weights[nn.NbLayers+1][input])
	}
}

func (nn *NeuralNetwork) GetWeights() []int64 {
	weights := make([]int64, 0)

	return weights
}

// TODO generalize
func (nn *NeuralNetwork) Forward(input []float64) ([][]float64, error) {
	if len(input) != nn.InputDimensions {
		return nil, errors.New("[neural.Forward]: input dimensions don't match the input layer")
	}

	stepsOutputs := make([][]float64, 0)

	z1 := make([]float64, 5)
	for i, xi := range input {
		z1[0] += xi * nn.Weights[0][i][0]
		z1[1] += xi * nn.Weights[0][i][1]
		z1[2] += xi * nn.Weights[0][i][2]
		z1[3] += xi * nn.Weights[0][i][3]
		z1[4] += xi * nn.Weights[0][i][4]
	}
	stepsOutputs = append(stepsOutputs, z1)
	x1 := Sigmoid(z1)

	z2 := make([]float64, 1)
	for i, xi := range x1 {
		z2[0] += xi * nn.Weights[2][0][i]
	}
	stepsOutputs = append(stepsOutputs, z2)
	output := Sigmoid(z2)

	stepsOutputs = append(stepsOutputs, output)
	return stepsOutputs, nil
}

// TODO generalize
func (nn *NeuralNetwork) Backpropagation(input []float64, output float64) ([][]float64, error) {
	steps, err := nn.Forward(input) // steps contains z1, z2, output
	if err != nil {
		return nil, err
	}

	x1 := Sigmoid(steps[0])

	fmt.Println(steps)
	delta2 := make([]float64, len(steps[0])-1)
	for i := range steps[1] {
		delta2[i] = (steps[2][0] - output) * GradSigmoid(steps[1])[i]
	}

	delta_w2 := make([]float64, len(delta2))
	for i := range delta2 {
		delta_w2[i] = delta2[i] * x1[i]
	}

	delta1 := make([]float64, len(steps[0])-1)
	for i := range delta2 {
		for _, xi := range delta2 {
			delta1[i] += xi * nn.Weights[2][0][i] * GradSigmoid(steps[0])[i]
		}
	}

	delta_w1 := make([]float64, len(delta1))
	for i := range delta1 {
		delta_w1[i] = delta1[i] * input[i]
	}

	gradients := [][]float64{delta_w2, delta_w1}
	fmt.Println(delta_w1)
	fmt.Println(delta_w2)

	return gradients, nil
}

// -------------- UTILS

func Sigmoid(x []float64) []float64 {
	var x2 []float64
	for _, xi := range x {
		x2 = append(x2, 1/(1+math.Exp(-xi)))
	}
	return x2
}

func GradSigmoid(x []float64) []float64 {
	var x2 []float64
	for _, xi := range x {
		mxi := []float64{xi}
		x2 = append(x2, Sigmoid(mxi)[0]*(1-Sigmoid(mxi)[0]))
	}
	return x2
}

func RandomArrayFloat(size int) []float64 {
	randomArray := make([]float64, size)
	for i := range randomArray {
		randomArray[i] = rand.Float64()
	}
	return randomArray
}
