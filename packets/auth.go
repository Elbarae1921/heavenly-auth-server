package packets

// From Client
type LoginMessage struct {
   Username string
   Password string
}

// From Client
type RegisterMessage struct {
   Username string
   Password string
   Email    string
}

type Token struct {
   UserId    uint64
   ExpiresAt int64
}

// From Server
type LoginResponse struct {
   Token     Token
   Signature []byte
}

// From Server
type Error struct {
   Message string
}

type HandshakeMessage struct {
   Token     Token
   Signature []byte
}
