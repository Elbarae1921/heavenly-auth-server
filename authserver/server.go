package authserver

import (
	"encoding/gob"
	"fmt"
	"log"
	"main/authservice"
	"main/db"
	"main/gmessages"
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
		var _ gmessages.LoginMessage
		// check if our packet id is a proper key for our map
		if packetString, ok := packets.PACKETS_INT_TO_STRING[packetId]; ok {
			log.Printf("Received %s\n", packetString)
			// get the interface associated with the packetString
			if packetInterface, ok := packets.PACKET_TO_GAME_MESSAGE[packetString]; ok {
				var message = reflect.New(reflect.TypeOf(packetInterface)).Interface()

				err := msgpack.Unmarshal(packet.Content, &message)

				if err != nil {
					log.Println("Error unmarshalling packet: ", err)
				}

				logic <- packets.PacketDTO{
					Source:       packet.Source,
					Message:      message,
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
		packetString := packet.PacketString

		method := reflect.ValueOf(as).MethodByName(packetString + "Handle")
		fmt.Println(&method)
		if method.IsValid() {
			// get the value of packet.Message
			message := reflect.ValueOf(packet.Message).Interface()
			log.Printf("Calling method %sHandle\n with interface value of %s", packetString, message)

			// get return value of meth.Call
			ret := method.Call([]reflect.Value{reflect.ValueOf(packet.Message)})

			retVal := ret[0]

			log.Println(retVal)
		}

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
