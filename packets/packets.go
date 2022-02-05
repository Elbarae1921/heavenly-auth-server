package packets

import (
	"main/gmessages"
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

type ReturnPacket struct {
	ID      uint16
	Content []byte
}

type TokenPacketContent struct {
	Token     Token
	Signature []byte
}

var PACKETS_INT_TO_STRING = map[uint16]string{
	101: "MSG_LOGIN",
	102: "MSG_REGISTER",
}

var PACKETS_STRING_TO_INT = map[string]uint16{
	"MSG_LOGIN":    101,
	"MSG_REGISTER": 102,
}

var PACKET_TO_GAME_MESSAGE = map[string]interface{}{
	"MSG_LOGIN":    gmessages.LoginMessage{},
	"MSG_REGISTER": gmessages.RegisterMessage{},
}
