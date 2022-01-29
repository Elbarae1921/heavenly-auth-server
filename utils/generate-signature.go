package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"log"
	"main/packets"

	"github.com/vmihailenco/msgpack/v5"
)

func GenerateSignature(token packets.Token, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	tokenBytes, err := msgpack.Marshal(token)
	if err != nil {
		log.Println("Error occurred while marshalling token: ", err)
		return nil, err
	}

	mac.Write(tokenBytes)
	return mac.Sum(nil), nil
}
