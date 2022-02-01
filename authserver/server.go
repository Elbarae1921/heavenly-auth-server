package authserver

import (
	"bytes"
	"io"
	"log"
	"main/authservice"
	"main/db"
	"main/packets"
	"main/responses"
	"main/utils"
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
func (as *AuthServer) InitializeChannels() (chan<- packets.PacketWithConnDTO, chan<- packets.ReturnPacketDTO) {
	log.Println("Initializing channels")
	receive := make(chan packets.PacketWithConnDTO, 10)
	send := make(chan packets.ReturnPacketDTO, 10)
	logic := make(chan packets.PacketDTO, 10)
	go as.handleWriteChannel(send)
	go as.handleReadChannel(receive, logic)
	go as.handleLogicChannel(logic, send)

	return receive, send
}

func (as *AuthServer) handleReadChannel(receive <-chan packets.PacketWithConnDTO, logic chan<- packets.PacketDTO) {
	log.Println("Initialized read channel")
	for {
		packet := <-receive
		packetId := packet.ID
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

func (as *AuthServer) handleWriteChannel(send <-chan packets.ReturnPacketDTO) {
	log.Println("Initialized write channels")
	for {
		packet := <-send

		var buf bytes.Buffer

		enc := msgpack.NewEncoder(&buf)
		enc.UseArrayEncodedStructs(true)
		err := enc.Encode(&packets.Packet{ID: packet.ID, Content: packet.Content})
		if err != nil {
			panic(err)
		}

		log.Println("Sending bytes: ", buf.Bytes())
		packet.Conn.Write(buf.Bytes())
	}
}

//$2a$10$KAene/IdZcDp4szlMghh5OySlP9oCdqbALGIEBqNPdBJ5G7vajYri
func (as *AuthServer) handleLogicChannel(logic <-chan packets.PacketDTO, send chan<- packets.ReturnPacketDTO) {
	log.Println("Initialized logic channel")
	for {
		packet := <-logic
		packetString := packet.PacketString

		method := reflect.ValueOf(as).MethodByName(packetString + "Handle")
		if method.IsValid() {
			// get return value of method.Call
			ret := method.Call([]reflect.Value{reflect.ValueOf(packet.Message)})

			retVal, err := ret[0], ret[1]

			// var content []byte
			// var encodeError error

			// get reflect of err and check if it's nil
			if err.Interface() != nil {
				// get the error string from the error interface
				errString := err.Interface().(error).Error()

				var content []byte

				if response, encodeError := utils.Encode(responses.Fail(errString)); encodeError == nil {
					content = response
				} else {
					log.Println("Error encoding response: ", encodeError)
					content = []byte(errString)
				}

				send <- packets.ReturnPacketDTO{
					ID:      packet.ID,
					Conn:    packet.Connection,
					Content: content,
				}
				continue
			}

			var content []byte

			log.Printf("%v\n", retVal.Interface().([]byte))
			Response := responses.Success(retVal.Interface().([]byte))

			if response, encodeError := utils.Encode(Response); encodeError == nil {
				content = response
			} else {
				log.Println("Error encoding response: ", encodeError)
				content = []byte(err.Interface().(error).Error())
			}

			send <- packets.ReturnPacketDTO{
				ID:      packet.ID,
				Conn:    packet.Connection,
				Content: content,
			}
		}
	}
}

func (as *AuthServer) ListenForPackets(conn net.Conn, receive chan<- packets.PacketWithConnDTO) {
	log.Println("Listening for packets from connection: ", conn.RemoteAddr())
	buffer := make([]byte, 1024*4)
	for {
		// read the data into the buffer
		_, err := conn.Read(buffer)

		if err != nil {
			log.Println("Error reading from connection: ", err)
			return
		}

		var packet packets.Packet

		err = msgpack.Unmarshal(buffer, &packet)
		if err != nil {
			log.Println("Error unmarshalling packet: ", err)
			conn.Close()
			continue
		}

		// create a new packet with conn
		packetWithConn := packets.PacketWithConnDTO{
			ID:      packet.ID,
			Content: packet.Content,
			Conn:    conn,
		}

		receive <- packetWithConn
	}

}
