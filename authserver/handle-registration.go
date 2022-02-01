package authserver

import (
	"bytes"
	"main/gmessages"
	"main/packets"
	"main/utils"

	"github.com/vmihailenco/msgpack/v5"
)

func (as *AuthServer) MSG_REGISTERHandle(data interface{}) ([]byte, error) {
	newData := *data.(*gmessages.RegisterMessage)

	token, err := as.AuthService.Register(newData.UserName, newData.Email, newData.Password)
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

	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	enc.UseArrayEncodedStructs(true)
	if err := enc.Encode(&tokenPack); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
