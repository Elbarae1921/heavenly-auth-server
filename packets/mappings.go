package packets

var PACKETS_INT_TO_STRING = map[uint16]string{
  101: "LoginMessage",
  102: "RegisterMessage",
  1001: "HandshakeMessage",
  1002: "ChatMessage"
}

var PACKETS_STRING_TO_INT = map[string]uint16{
  "LoginMessage": 101,
  "RegisterMessage": 102,
  "HandshakeMessage": 1001,
  "ChatMessage": 1002
}

var PACKET_TO_GAME_MESSAGE = map[string]interface{}{
  "LoginMessage": LoginMessage{},
  "RegisterMessage": RegisterMessage{},
  "HandshakeMessage": HandshakeMessage{},
  "ChatMessage": ChatMessage{}
}
