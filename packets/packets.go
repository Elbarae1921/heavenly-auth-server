package packets

import "main/gmessages"

type Packet struct {
	ID      uint8
	Content []byte
	Source  string
}

type PacketDTO struct {
	Message      interface{}
	Source       string
	PacketString string
}

type TokenPacket struct {
	Token     Token
	Signature string
}

type Token struct {
	UserId    uint64
	ExpiresAt int64
}

var PACKETS_INT_TO_STRING = map[uint8]string{
	101: "MSG_LOGIN",
}

var PACKETS_STRING_TO_INT = map[string]uint8{
	"MSG_LOGIN": 101,
}

var PACKET_TO_GAME_MESSAGE = map[string]interface{}{
	"MSG_LOGIN": gmessages.LoginMessage{},
}
