package models


type User struct{
	Username string
	FullName string
	Email string
	PasswordHash string
	IsDisables bool
	JsonToken string
}

type Configuration struct {
	DbCreds string
	Secret string
}
