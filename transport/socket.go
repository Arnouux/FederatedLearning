package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
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
	return Socket{connection: pktConn, chanAck: make(chan bool)}, nil
}

func (s *Socket) Send(dest string, pkt Packet) error {
	addr, err := net.ResolveUDPAddr("udp", dest)
	if err != nil {
		return err
	}

	// TODO separate different packet types

	// Decompose packet
	sendingPkt := Packet{
		Source:      pkt.Source,
		Destination: pkt.Destination,
		Type:        pkt.Type,
	}

	i := 50000
	for i < len(pkt.Message) {
		sendingPkt.Message = pkt.Message[i-50000 : i]

		// Marshal and send current packet
		bytes, err := json.Marshal(sendingPkt)
		if err != nil {
			return err
		}
		n, err := s.connection.WriteTo(bytes, addr)
		if err != nil {
			return err
		}
		fmt.Println("sent")
		if n < len(bytes) {
			return errors.New("[transport.Socket.Send]: not all bytes were written")
		}

		i += 50000

		// waits for ack before sending next packet
		fmt.Println("waiting for ack")
		<-s.chanAck
		fmt.Println("ack received")
	}
	// Marshal and send last packet
	sendingPkt.Message = pkt.Message[i-50000:]
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

	// bytes, err := json.Marshal(pkt)
	// if err != nil {
	// 	return err
	// }

	// n, err := s.connection.WriteTo(bytes, addr)
	// if err != nil {
	// 	return err
	// }
	// if n < len(bytes) {
	// 	return errors.New("[transport.Socket.Send]: not all bytes were written")
	// }

	return nil
}

func (s *Socket) Recv() (Packet, error) {
	messageFull := ""

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

	fmt.Println("switch type", pkt.Type)
	switch pkt.Type {
	case Ack:
		s.chanAck <- true
	case EncryptedChunk:
		count := 0
		// send ack
		pktAck := Packet{
			Source:      s.address,
			Destination: pkt.Source,
			Message:     string(count),
			Type:        Ack,
		}
		count += 1
		s.Send(pkt.Source, pktAck)

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
			fmt.Println("recv")

			// send ack
			pktAck := Packet{
				Source:      s.address,
				Destination: pkt.Source,
				Message:     string(count),
				Type:        Ack,
			}
			count += 1
			s.Send(pkt.Source, pktAck)
			fmt.Println("ack sent to source")

			// TODO use terminal token
			if len(pkt.Message) != 50000 {
				// TODO make sure every packets are from same process (dest/source)
				break
			}
		}
	}

	pktFinal := Packet{
		Source:      pkt.Source,
		Destination: pkt.Destination,
		Message:     messageFull,
		Type:        pkt.Type,
	}
	return pktFinal, nil
}

func (s *Socket) GetAdress() string {
	return s.connection.LocalAddr().String()
}
