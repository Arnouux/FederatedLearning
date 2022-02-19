package transport

type Packet struct {
	Source      string
	Destination string
	Message     string
	Type        string
	ID          string
}

const (
	EncryptedChunk = "encryptedChunk"
	Ack            = "acknowlegdement"
	Result         = "result"
	Join           = "join"
)

func (p *Packet) String() string {
	return "[" + p.Source + "] -> [" + p.Destination + "], message=\"" + p.Message + "\""
}
