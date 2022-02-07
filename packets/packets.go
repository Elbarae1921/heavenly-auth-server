package packets

import (
	"net"
)

type PacketWithConnDTO struct {
	ID      uint16
	Content []byte
	Conn    net.Conn
}

type PacketDTO struct {
	ID           uint16
	Message      interface{}
	Connection   net.Conn
	PacketString string
}

type ReturnPacketDTO struct {
	ID      uint16
	Content []byte
	Conn    net.Conn
}

type TokenPacketContent struct {
	Token     Token
	Signature []byte
}
