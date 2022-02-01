package authserver

import (
	"main/packets"
)

func (as *AuthServer) MSG_REGISTERHandle(data interface{}) (*packets.TokenPacketContent, error) {
	// newData := *data.(*gmessages.RegisterMessage)

	// token, err := as.AuthService.Register(newData.UserName, newData.Email, newData.Password)
	// if err != nil {
	// 	return nil, err
	// }

	// // sign the token
	// signature, err := utils.GenerateSignature(*token, key)
	// if err != nil {
	// 	return nil, err
	// }

	// // msg pack the token and signature
	// tokenPack := &packets.TokenPacketContent{
	// 	Token:     *token,
	// 	Signature: signature,
	// }

	// return tokenPack, nil
	return nil, nil
}
