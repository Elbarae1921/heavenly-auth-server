package packets

// From Client
type MoveMessage struct {
   X float32
   Y float32
}

// From Server
type MovedMessage struct {
   Id uint64
   X  float32
   Y  float32
}

// From Server
type MovableSpawned struct {
   Id uint64
   X  float32
   Y  float32
}
