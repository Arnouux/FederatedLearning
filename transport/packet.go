package transport

import "encoding/json"

type Packet struct {
	Destination string
	Message     string
	Payload     json.RawMessage
}
