package transport

type Packet struct {
	Source      string
	Destination string
	Message     string
	Type        string
}

const (
	EncryptedChunk = "encryptedChunk"
	Ack            = "acknowlegdement"
	Result         = "result"
)

func (p *Packet) String() string {
	return "[" + p.Source + "] -> [" + p.Destination + "], message=\"" + p.Message + "\""
}
