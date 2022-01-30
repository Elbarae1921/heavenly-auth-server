package authserver

import (
	"encoding/gob"
	"fmt"
	"log"
	"main/authservice"
	"main/db"
	"main/packets"
	"net"
	"reflect"

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
	logic := make(chan packets.PacketDTO, 10)
	go as.handleWriteChannel(send)
	go as.handleReadChannel(receive, logic)
	go as.handleLogicChannel(logic, send)

	return receive, send
}

func (as *AuthServer) handleReadChannel(receive <-chan packets.Packet, logic chan<- packets.PacketDTO) {
	log.Println("Handling read channel")
	for {
		packet := <-receive
		packetId := packet.ID
		// check if our packet id is a proper key for our map
		if packetString, ok := packets.PACKETS_INT_TO_STRING[packetId]; ok {
			log.Printf("Received %s\n", packetString)
			// get the interface associated with the packetString
			if packetInterface, ok := packets.PACKET_TO_GAME_MESSAGE[packetString]; ok {
				var message = reflect.New(reflect.TypeOf(packetInterface)).Interface()
				log.Println("Message: ", message)
				// get the type of message interface
				messageType := reflect.TypeOf(message)
				log.Println("Message type: ", messageType)
				err := msgpack.Unmarshal(packet.Content, &messageType)

				if err != nil {
					log.Println("Error unmarshalling packet: ", err)
					// panic(err)
					//TODO close connection
				}

				logic <- packets.PacketDTO{
					Source:       packet.Source,
					Message:      &messageType,
					PacketString: packetString,
				}

			}

		}

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

func (as *AuthServer) handleLogicChannel(logic <-chan packets.PacketDTO, send chan<- packets.Packet) {
	log.Println("Handling logic channel")
	for {
		packet := <-logic

		// get our packet string
		packetString := packet.PacketString
		// get our source connection
		// source := packet.Source

		meth := reflect.ValueOf(as).MethodByName(packetString + "Handle")
		fmt.Println(meth)
		// if meth.IsValid() {
		// meth.Call([]reflect.Value{reflect.ValueOf(as.AuthService), reflect.ValueOf(source)})
		// }

		// if packetGameMessage, ok := packet.Message.(*gmessages.LoginMessage); ok {
		// 	log.Println("Received login message")

		// 	// // handle login
		// 	_, err := as.HandleLogin(*packetGameMessage)
		// 	if err != nil {
		// 		// validate error here
		// 	}
		// 	send <- packets.PacketDTO{
		// 		Source:  packetGameMessage.UserIp,
		// 		Content: token,
		// 	}
		// }
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
		packet.Source = conn.RemoteAddr().String()
		if err != nil {
			log.Println("Error unmarshalling packet: ", err)
			// panic(err)
		}

		receive <- packet
	}
}
