package transport

type Packet struct {
	Source      string
	Destination string
	Message     string
	Type        string
	ID          string
	Params      Parameters
}

type Parameters struct {
	InputDimensions    int
	OutputDimensions   int
	NbLayers           int
	NbNeurons          int
	LearningRate       float64
	NbIterations       int
	ActivationFunction string
	BatchSize          int
}

const (
	EncryptedChunk = "encryptedChunk"
	Ack            = "acknowlegdement"
	Result         = "result"
	Join           = "join"
	Params         = "params"
)

func (p *Packet) String() string {
	return "[" + p.Source + "] -> [" + p.Destination + "], message=\"" + p.Message + "\""
}
