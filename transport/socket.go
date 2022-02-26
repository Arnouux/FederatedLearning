package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
)

type Socket struct {
	address    string
	connection net.PacketConn
	chanAck    chan bool
}

func CreateSocket() (Socket, error) {
	pktConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return Socket{}, err
	}
	return Socket{connection: pktConn, chanAck: make(chan bool, 1)}, nil
}

func (s *Socket) Send(dest string, pkt Packet) error {
	addr, err := net.ResolveUDPAddr("udp", dest)
	if err != nil {
		return err
	}

	// Decompose packet
	sendingPkt := Packet{
		Source:      pkt.Source,
		Destination: pkt.Destination,
		Type:        pkt.Type,
	}

	// Prepare pkt according to type
	switch pkt.Type {
	case Ack:
		sendingPkt.Message = pkt.Message
	case EncryptedChunk, Result:
		i := 50000
		for i < len(pkt.Message) {
			sendingPkt.Message = pkt.Message[i-50000 : i]

			// Marshal and send current packet
			bytes, err := json.Marshal(sendingPkt)
			if err != nil {
				fmt.Println(err)
				return err
			}
			n, err := s.connection.WriteTo(bytes, addr)
			if err != nil {
				fmt.Println(err)
				return err
			}
			if n < len(bytes) {
				fmt.Println("[transport.Socket.Send]: not all bytes were written")
				return errors.New("[transport.Socket.Send]: not all bytes were written")
			}

			i += 50000

			// waits for ack before sending next packet
			//time.Sleep(time.Millisecond * 10)
			<-s.chanAck
			// fmt.Println(b)
		}
		// Marshal and send last packet
		sendingPkt.Message = pkt.Message[i-50000:]

	default:
		sendingPkt.Message = pkt.Message
	}

	bytes, err := json.Marshal(sendingPkt)
	if err != nil {
		return err
	}
	n, err := s.connection.WriteTo(bytes, addr)
	if err != nil {
		return err
	}
	if n < len(bytes) {
		return errors.New("[transport.Socket.Send]: not all bytes were written")
	}
	return nil
}

func (s *Socket) Recv() (Packet, error) {
	var messageFull string

	buffer := make([]byte, 65000)
	n, _, err := s.connection.ReadFrom(buffer)
	if err != nil {
		fmt.Println(err)
		return Packet{}, err
	}
	pkt := Packet{}
	err = json.Unmarshal(buffer[0:n], &pkt)
	if err != nil {
		fmt.Println(err)
		return Packet{}, err
	}
	messageFull += pkt.Message

	switch pkt.Type {
	case Ack:
		s.chanAck <- true
	case EncryptedChunk, Result:
		count := 0
		// send ack
		pktAck := Packet{
			Source:      s.GetAddress(),
			Destination: pkt.Source,
			Message:     strconv.Itoa(count),
			Type:        Ack,
		}
		count += 1
		s.Send(pkt.Source, pktAck)

		currentSource := pkt.Source

		// Waiting for other packets
		for {
			// Reads up to 65000
			// -> Sender needs to wait for acknowlegdement before sending next packets
			buffer := make([]byte, 65000)
			n, _, err := s.connection.ReadFrom(buffer)
			if err != nil {
				fmt.Println(err)
				return Packet{}, err
			}
			pkt := Packet{}
			err = json.Unmarshal(buffer[0:n], &pkt)
			if err != nil {
				fmt.Println(err)
				return Packet{}, err
			}
			messageFull += pkt.Message

			// Check pkt is from the same message
			if pkt.Source != currentSource || pkt.Type != EncryptedChunk {
				// TODO : start new recv process or ignore ?
				// TODO : add rng ID to make sure it is from unique message
			}

			// send ack
			pktAck := Packet{
				Source:      s.GetAddress(),
				Destination: pkt.Source,
				Message:     strconv.Itoa(count),
				Type:        Ack,
			}
			count += 1
			s.Send(pkt.Source, pktAck)

			// TODO use terminal token or 'end' field in packet
			if len(pkt.Message) != 50000 {
				break
			}
		}
	default:

	}

	pktFinal := Packet{
		Source:      pkt.Source,
		Destination: pkt.Destination,
		Message:     messageFull,
		Type:        pkt.Type,
	}
	return pktFinal, nil
}

func (s *Socket) GetAddress() string {
	return s.connection.LocalAddr().String()
}
