package authservice

import (
	"context"
	"errors"
	"log"
	"main/db"
	"main/packets"
	"main/servererrors"
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
		return nil, errors.New(servererrors.ErrInvalid)
	} else if err != nil {
		log.Printf("Error occurred: %s", err)
		return nil, errors.New(servererrors.ErrInternal)
	}

	// check if the password matches with bcrypt CompareHashAndPassword
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), passwordBytes); err != nil {
		return nil, errors.New(servererrors.ErrInvalid)
	}
	// create a timestamp that expires in 5 minutes
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	return &packets.Token{
		UserId:    uint64(account.ID),
		ExpiresAt: expiresAt,
	}, nil
}

func (as *AuthService) Register(username string, email string, password string) (*packets.Token, error) {

	// convert the password to a byte array
	passwordBytes := []byte(password)

	ctx := context.Background()

	// find the username in the database
	account, err := as.Client.Account.FindFirst(db.Account.Or(
		db.Account.Username.Equals(username),
		db.Account.Email.Equals(email),
		db.Account.Username.Equals(email),
		db.Account.Email.Equals(username),
	)).Exec(ctx)

	if err != nil && err != db.ErrNotFound {
		log.Printf("Error occurred: %s", err)
		return nil, errors.New(servererrors.ErrInternal)
	}

	// check if the username already exists
	if account != nil {
		return nil, errors.New(servererrors.ErrExists)
	}

	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)

	if err != nil {
		log.Printf("Error occurred: %s", err)
		return nil, errors.New(servererrors.ErrInternal)
	}

	// create a new account
	createdAccount, err := as.Client.Account.CreateOne(
		db.Account.Username.Set(username),
		db.Account.Password.Set(string(hashedPassword)),
		db.Account.Email.Set(email),
	).Exec(ctx)
	if err != nil {
		return nil, errors.New(servererrors.ErrInternal)
	}

	// create a timestamp that expires in 5 minutes
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	return &packets.Token{
		UserId:    uint64(createdAccount.ID),
		ExpiresAt: expiresAt,
	}, nil
}
