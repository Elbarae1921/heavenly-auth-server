package utils

import (
	"bufio"
	"log"
	"os"
)

// generate a function that loads the rsa.private file and skips the first line
// this is because the first line is the comment
func LoadPrivateKey() ([]byte, error) {
	var privateKey string
	// find the private key file
	privateKeyFile, err := os.Open("./rsa.private")
	if err != nil {
		log.Fatalln("Couldn't load private key file")
	}

	scanner := bufio.NewScanner(privateKeyFile)
	scanner.Scan() // this moves to the next line

	for scanner.Scan() {
		// check if the line contains the - character, if it does we've reached the end of the key and should stop
		if scanner.Text() == "-" {
			break
		}

		privateKey += scanner.Text()
	}

	return []byte(privateKey), nil
}
