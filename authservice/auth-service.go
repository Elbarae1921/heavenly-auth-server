package authservice

import (
	"context"
	"errors"
	"log"
	"main/db"
	"main/packets"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Client *db.PrismaClient
}

func (as *AuthService) Login(username string, password string) (*packets.Token, error) {

	// convert the password to a byte array
	passwordBytes := []byte(password)

	ctx := context.Background()

	// find the username in the database
	account, err := as.Client.Account.FindUnique(
		db.Account.Username.Equals(username),
	).Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		log.Printf("No record found for username: %s", username)
		return nil, errors.New("No record found for username")
	} else if err != nil {
		log.Printf("Error occurred: %s", err)
		return nil, err
	}

	// check if the password matches with bcrypt CompareHashAndPassword
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), passwordBytes); err != nil {
		log.Printf("Attempted password: %s did not match account password", password)
		return nil, err
	}
	// create a timestamp that expires in 5 minutes
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	return &packets.Token{
		UserId:    uint64(account.ID),
		ExpiresAt: expiresAt,
	}, nil
}
