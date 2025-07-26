package users

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/umaaamm/contact/mongo"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"name"`
	Password string `json:"password"`
}

func (user *User) Create() (primitive.ObjectID, error) {
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return primitive.NilObjectID, err
	}
	user.Password = hashedPassword

	user.ID = primitive.NewObjectID().Hex()

	res, err := mongo.DB.Collection("user").InsertOne(context.TODO(), user)
	if err != nil {
		return primitive.NilObjectID, err
	}

	id := res.InsertedID.(primitive.ObjectID)
	return id, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetUserIdByUsername(username string) (string, error) {
	var user User

	err := mongo.DB.Collection("user").FindOne(
		context.TODO(),
		bson.M{"username": username},
	).Decode(&user)

	if err != nil {
		return "", err
	}

	if user.ID == primitive.NilObjectID.String() {
		return "", errors.New("user not found")
	}

	return user.ID, nil
}

func (user *User) Authenticate() bool {
	var dbUser User

	err := mongo.DB.Collection("user").FindOne(
		context.TODO(),
		bson.M{"username": user.Username},
	).Decode(&dbUser)

	if err != nil {
		if err == mongoDriver.ErrNoDocuments {
			return false
		}
		log.Fatal(err)
	}

	return CheckPasswordHash(user.Password, dbUser.Password)
}
