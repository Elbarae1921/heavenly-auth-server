package main

import (
	"log"
	"main/authserver"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// initialize our server struct
	as, err := authserver.Start(os.Getenv("IP"), os.Getenv("PORT"))

	if err != nil {
		log.Fatal(err)
	}

	receive, _ := as.InitializeChannels()

	for {
		// create buffer
		conn, err := as.N_TCPListen.Accept()
		log.Println("Received connection from ", conn.RemoteAddr())
		if err != nil {
			continue
		}

		go as.ListenForPackets(conn, receive)

	}

}
