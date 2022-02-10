package transport

import (
	"encoding/json"
	"errors"
	"net"
)

type Socket struct {
	address    string
	connection net.PacketConn
}

func CreateSocket() (Socket, error) {
	pktConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return Socket{}, err
	}
	return Socket{connection: pktConn}, nil
}

func (s *Socket) Send(dest string, pkt Packet) error {
	addr, err := net.ResolveUDPAddr("udp", dest)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(pkt)
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
	buffer := make([]byte, 65000)
	n, _, err := s.connection.ReadFrom(buffer)
	if err != nil {
		return Packet{}, err
	}
	pkt := Packet{}
	json.Unmarshal(buffer[0:n], &pkt)
	return pkt, nil
}

func (s *Socket) GetAdress() string {
	return s.connection.LocalAddr().String()
}
