package transport

type Packet struct {
	Source      string
	Destination string
	Message     string
}

func (p *Packet) String() string {
	return "[" + p.Source + "] -> [" + p.Destination + "], message=\"" + p.Message + "\""
}
