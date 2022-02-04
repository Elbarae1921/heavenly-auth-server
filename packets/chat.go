package packets

// From Client
type ChatMessage struct {
   Text string
}

// From Server
type ExternalChatMessage struct {
   Author string
   Text   string
}
