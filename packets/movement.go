package packets

// From Client
type MoveMessage struct {
   X int64
   Y int64
}

// From Server
type MoveCorrectionMessage struct {
   X int64
   Y int64
}
