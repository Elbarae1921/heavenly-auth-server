package authserver

import (
	"errors"
	"main/gmessages"
	"main/packets"
	"main/utils"
)

func (as *AuthServer) MSG_LOGINHandle(data interface{}) ([]byte, error) {
	newData := *data.(*gmessages.LoginMessage)
	// load our private key
	key, err := utils.LoadPrivateKey()
	// generate a new token
	if err != nil {
		return nil, err
	}

	token, err := as.AuthService.Login(newData.UserName, newData.Password)
	if err != nil {
		return nil, err
	}

	// sign the token
	signature, err := utils.GenerateSignature(*token, key)
	if err != nil {
		return nil, err
	}

	// msg pack the token and signature
	tokenPack := &packets.TokenPacketContent{
		Token:     *token,
		Signature: signature,
	}

	// msg pack the token

	bytes, err := utils.Encode(&tokenPack)

	if err != nil {
		return nil, errors.New("INTERNAL_ERROR")
	}

	return bytes, nil
}
