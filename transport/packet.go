package transport

import (
	"encoding/json"
)

type Packet struct {
	Source      string
	Destination string
	Message     string
	Payload     json.RawMessage
}

func (p *Packet) String() string {
	return "[" + p.Source + "] -> [" + p.Destination + "], message=\"" + p.Message + "\""
}
