package authserver

import (
	"main/gmessages"
	"main/packets"
	"main/utils"

	"github.com/vmihailenco/msgpack/v5"
)

func (as *AuthServer) HandleLogin(data gmessages.LoginMessage) ([]byte, error) {
	// load our private key
	key, err := utils.LoadPrivateKey()
	// generate a new token
	if err != nil {
		return nil, err
	}

	token, err := as.AuthService.Login(data.UserName, data.Password)
	if err != nil {
		return nil, err
	}

	// sign the token
	signature, err := utils.GenerateSignature(*token, key)
	if err != nil {
		return nil, err
	}

	// msg pack the token and signature
	tokenPack := &packets.TokenPacket{
		Token:     *token,
		Signature: string(signature),
	}

	// msg pack the token
	tokenPackBytes, err := msgpack.Marshal(tokenPack)
	if err != nil {
		return nil, err
	}

	return tokenPackBytes, nil
}
