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

	// check if rsa.private file exists, if it doesn't, log fatal
	_, err = os.Stat("./rsa.private")
	if os.IsNotExist(err) {
		log.Fatal("Error: rsa.private file not found")
	}

}

func main() {
	// initialize our server struct
	as, err := authserver.Start(os.Getenv("IP"), os.Getenv("PORT"))

	if err != nil {
		log.Fatal(err)
	}

	receive, _ := as.InitializeChannels()
	defer func() {
		if err := as.AuthService.Client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	for {
		conn, err := as.TCPListen.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go as.ListenForPackets(conn, receive)

	}

}
