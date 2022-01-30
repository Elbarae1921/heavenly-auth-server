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
func (as *AuthServer) InitializeChannels() (chan<- packets.PacketWithConn, chan<- packets.ReturnPacket) {
	log.Println("Initializing channels")
	receive := make(chan packets.PacketWithConn, 10)
	send := make(chan packets.ReturnPacket, 10)
	logic := make(chan packets.PacketDTO, 10)
	go as.handleWriteChannel(send)
	go as.handleReadChannel(receive, logic)
	go as.handleLogicChannel(logic, send)

	return receive, send
}

func (as *AuthServer) handleReadChannel(receive <-chan packets.PacketWithConn, logic chan<- packets.PacketDTO) {
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
					packet.Conn.Close()
					continue
				}

				logic <- packets.PacketDTO{
					ID:           packetId,
					Connection:   packet.Conn,
					Message:      message,
					PacketString: packetString,
				}

			}

		}

	}
}

func (as *AuthServer) handleWriteChannel(send <-chan packets.ReturnPacket) {
	log.Println("Initialized write channels")
	for {
		packet := <-send
		packet.Conn.Write(packet.Content)
	}
}

//$2a$10$KAene/IdZcDp4szlMghh5OySlP9oCdqbALGIEBqNPdBJ5G7vajYri
func (as *AuthServer) handleLogicChannel(logic <-chan packets.PacketDTO, send chan<- packets.ReturnPacket) {
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
				send <- packets.ReturnPacket{
					ID:      packet.ID,
					Conn:    packet.Connection,
					Content: []byte(errString),
				}
				continue
			}

			send <- packets.ReturnPacket{
				ID:      packet.ID,
				Conn:    packet.Connection,
				Content: retVal.Interface().([]byte),
			}
		}
	}
}

func (as *AuthServer) ListenForPackets(conn net.Conn, receive chan<- packets.PacketWithConn) {
	log.Println("Listening for packets from connection: ", conn.RemoteAddr())
	buffer := make([]byte, 1024*4)
	for {
		// read the data into the buffer
		length, err := conn.Read(buffer)

		if err != nil {
			conn.Close()
			continue
		}

		// check if length is over 0, if it is, we have data, if not, continue to next iteration
		if length > 0 {

			log.Println("Read ", length, " bytes")

			//log receive connection
			// log.Println("Received packet from ", conn.RemoteAddr())

			var packet packets.Packet

			err = msgpack.Unmarshal(buffer, &packet)
			if err != nil {
				log.Println("Error unmarshalling packet: ", err)
				conn.Close()
				continue
			}

			// create a new packet with conn
			packetWithConn := packets.PacketWithConn{
				ID:      packet.ID,
				Content: packet.Content,
				Conn:    conn,
			}

			receive <- packetWithConn
		} else {
			continue
		}
	}
}
