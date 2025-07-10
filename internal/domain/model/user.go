package model

type User struct {
	Email        string
	PasswordHash string
	Role         string
	Version      float64
}
