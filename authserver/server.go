package authserver

import (
	"io"
	"log"
	"main/authservice"
	"main/db"
	"main/gmessages"
	"main/packets"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

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

	as.initializeLogFile()

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

func (as *AuthServer) initializeLogFile() {
	// get the folder path where main.go is located
	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// check if the logs folder exists
	_, err = os.Stat(appPath + "/logs")
	if os.IsNotExist(err) {
		// if it doesn't, create it
		err = os.Mkdir(appPath+"/logs", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// check if today's log file exists
	now := time.Now().Format("2006-01-02")
	filename := appPath + "/logs/authserver-" + now + ".log"

	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
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
	log.Println("Initialized read channel")
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
	log.Println("Initialized write channels")
	for {
		packet := <-send
		_, err := net.Dial("tcp", packet.Source)
		if err != nil {
			log.Println("Error dialing: ", err)
		}

		// encoder := gob.NewEncoder(conn)
		// encoder.Encode(packet.Content)
		// conn.Write(packet.Content)
	}
}

//$2a$10$KAene/IdZcDp4szlMghh5OySlP9oCdqbALGIEBqNPdBJ5G7vajYri
func (as *AuthServer) handleLogicChannel(logic <-chan packets.PacketDTO, send chan<- packets.Packet) {
	log.Println("Initialized logic channel")
	for {
		packet := <-logic
		packetString := packet.PacketString

		method := reflect.ValueOf(as).MethodByName(packetString + "Handle")
		if method.IsValid() {
			// get return value of method.Call
			ret := method.Call([]reflect.Value{reflect.ValueOf(packet.Message)})

			retVal, err := ret[0], ret[1]

			// get reflect of err and check if it's nil
			if err.Interface() != nil {
				// get the error string from the error interface
				errString := err.Interface().(error).Error()
				send <- packets.Packet{
					Source:  packet.Source + ":8888",
					Content: []byte(errString),
				}
				continue
			}

			send <- packets.Packet{
				Source:  packet.Source + ":8888",
				Content: retVal.Interface().([]byte),
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
		_, err := conn.Read(buffer)

		//log receive connection
		log.Println("Received packet from ", conn.RemoteAddr())

		if err != nil {
			continue
		}

		var packet packets.Packet

		err = msgpack.Unmarshal(buffer, &packet)
		// initialize a string that removes the : part of the ip address
		packet.Source = strings.Split(packet.Source, ":")[0]
		if err != nil {
			log.Println("Error unmarshalling packet: ", err)
			conn.Close()
		}

		receive <- packet
	}
}
