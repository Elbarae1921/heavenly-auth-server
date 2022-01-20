package authserver

import (
	"encoding/gob"
	"log"
	"main/gmessages"
	"main/packets"
	"net"

	"github.com/vmihailenco/msgpack/v5"
)

type AuthServer struct {
	ip          string
	port        string
	N_TCPListen *net.TCPListener
}

func Start(ip string, port string) (*AuthServer, error) {
	as := &AuthServer{
		ip:   ip,
		port: port,
	}

	if ok, err := as.listen(); !ok {
		return nil, err
	}

	return as, nil

}

func (as *AuthServer) listen() (bool, error) {
	localAddress, err := net.ResolveTCPAddr("tcp", as.ip+":"+as.port)
	if err != nil {
		return false, err
	}

	as.N_TCPListen, err = net.ListenTCP("tcp", localAddress)
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
			// handle login
			token, err := as.handleLogin(*packetGameMessage)
			if err != nil {
				// validate error here
			}
			send <- packets.Packet{
				Source:  packetGameMessage.ReturnAdr,
				Content: token,
			}
		}
	}
}

func (as *AuthServer) handleLogin(gmessages.LoginMessage) ([]byte, error) {
	// handle login
	return []byte("token"), nil
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
