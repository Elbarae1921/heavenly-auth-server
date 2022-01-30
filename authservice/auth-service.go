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

	// if username is test then get the encrypted password
	if username == "test" {
		password, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
		log.Println("password: ", string(password))
	}

	if errors.Is(err, db.ErrNotFound) {
		return nil, errors.New("invalid credentials")
	} else if err != nil {
		log.Printf("Error occurred: %s", err)
		return nil, errors.New("error occurred")
	}

	// check if the password matches with bcrypt CompareHashAndPassword
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), passwordBytes); err != nil {
		return nil, errors.New("invalid credentials")
	}
	// create a timestamp that expires in 5 minutes
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	return &packets.Token{
		UserId:    uint64(account.ID),
		ExpiresAt: expiresAt,
	}, nil
}
