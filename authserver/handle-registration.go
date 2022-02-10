package authserver

import (
	"errors"
	"main/packets"
	"main/utils"
)

func (as *AuthServer) RegisterMessageHandle(data interface{}) ([]byte, error) {
	newData := *data.(*packets.RegisterMessage)

	token, err := as.AuthService.Register(newData.Username, newData.Email, newData.Password)
	if err != nil {
		return nil, err
	}

	// msg pack the token and signature
	tokenPack := &packets.TokenPacketContent{
		Token: *token,
	}

	// msg pack the token

	bytes, err := utils.Encode(&tokenPack)

	if err != nil {
		return nil, errors.New("INTERNAL_ERROR")
	}

	return bytes, nil
}
