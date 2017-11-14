package Models

type User struct{
	Username string
	FullName string
	email string
	PasswordHash string
	PasswordSalt string
	IsDisables bool
}
