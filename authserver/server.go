package authserver

import (
	"encoding/gob"
	"log"
	"main/authservice"
	"main/db"
	"main/gmessages"
	"main/packets"
	"main/utils"
	"net"

	"github.com/vmihailenco/msgpack/v5"
)

type AuthServer struct {
	ip          string
	port        string
	TCPListen   *net.TCPListener
	AuthService *authservice.AuthService
}

func Start(ip string, port string) (*AuthServer, error) {
	as := &AuthServer{
		ip:   ip,
		port: port,
	}

	if ok, err := as.listen(); !ok {
		return nil, err
	}

	as.initializeAuthService()

	return as, nil

}

func (as *AuthServer) initializeAuthService() {
	as.AuthService = &authservice.AuthService{
		Client: db.NewClient(),
	}
	if err := as.AuthService.Client.Prisma.Connect(); err != nil {
		panic(err)
	}

	defer func() {
		if err := as.AuthService.Client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

}

func (as *AuthServer) listen() (bool, error) {
	localAddress, err := net.ResolveTCPAddr("tcp", as.ip+":"+as.port)
	if err != nil {
		return false, err
	}

	as.TCPListen, err = net.ListenTCP("tcp", localAddress)
	log.Println("Listening on ", as.ip+":"+as.port)

	if err != nil {
		return false, err
	}

	return true, nil
}

// Returns channels for reading and writing
func (as *AuthServer) InitializeChannels() (chan<- packets.Packet, chan<- packets.Packet) {
	log.Println("Initializing channels")
	receive := make(chan packets.Packet, 10)
	send := make(chan packets.Packet, 10)
	logic := make(chan interface{}, 40)
	go as.handleWriteChannel(send)
	go as.handleReadChannel(receive, logic)
	go as.handleLogicChannel(logic, send)

	return receive, send
}

func (as *AuthServer) handleReadChannel(receive <-chan packets.Packet, logic chan<- interface{}) {
	log.Println("Handling read channel")
	for {
		packet := <-receive
		log.Println("Received packet from ", packet.Source)
		packetId := packet.ID
		var loginMessage gmessages.LoginMessage
		// check if our packet is a valid one using our map
		if packetId == packets.PACKETS_STRING_TO_INT["MSG_LOGIN"] {
			// msgpack decode the packet
			err := msgpack.Unmarshal(packet.Content, &loginMessage)

			if err != nil {
				log.Println("Error unmarshalling packet: ", err)
				// panic(err)
			}
			// handle login
			logic <- &loginMessage
		}

		log.Println("Packet ID: ", packetId)
	}
}

func (as *AuthServer) handleWriteChannel(send <-chan packets.Packet) {
	log.Println("Handling write channels")
	for {
		packet := <-send
		conn, err := net.Dial("tcp", packet.Source)
		if err != nil {
			continue
		}
		encoder := gob.NewEncoder(conn)
		encoder.Encode(packet.Content)
		conn.Write(packet.Content)
	}
}

func (as *AuthServer) handleLogicChannel(logic <-chan interface{}, send chan<- packets.Packet) {
	log.Println("Handling logic channel")
	for {
		packet := <-logic
		if packetGameMessage, ok := packet.(*gmessages.LoginMessage); ok {
			log.Println("Received login message")

			// handle login
			token, err := as.handleLogin(*packetGameMessage)
			if err != nil {
				// validate error here
			}
			send <- packets.Packet{
				Source:  packetGameMessage.UserIp,
				Content: token,
			}
		}
	}
}

func (as *AuthServer) ListenForPackets(conn net.Conn, receive chan<- packets.Packet) {
	defer conn.Close()
	log.Println("Listening for packets")
	buffer := make([]byte, 1024*4)
	for {
		// read the data into the buffer
		length, err := conn.Read(buffer)

		log.Println("Received data from ", conn.RemoteAddr())
		log.Println("Length: ", length)
		//
		if err != nil {
			continue
		}

		var packet packets.Packet

		err = msgpack.Unmarshal(buffer, &packet)
		if err != nil {
			log.Println("Error unmarshalling packet: ", err)
			// panic(err)
		}

		receive <- packet
	}
}

func (as *AuthServer) handleLogin(data gmessages.LoginMessage) ([]byte, error) {
	// load our private key
	key, err := utils.LoadPrivateKey()
	// generate a new token
	if err != nil {
		return nil, err
	}

	token, err := as.AuthService.Login(data.UserName, data.Password)
	if err != nil {
		return nil, err
	}

	// sign the token
	signature, err := utils.GenerateSignature(*token, key)
	if err != nil {
		return nil, err
	}

	// msg pack the token and signature
	tokenPack := &packets.TokenPacket{
		Token:     *token,
		Signature: string(signature),
	}

	// msg pack the token
	tokenPackBytes, err := msgpack.Marshal(tokenPack)
	if err != nil {
		return nil, err
	}

	return tokenPackBytes, nil
}
