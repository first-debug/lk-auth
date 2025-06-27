// Доменная область
package models

type User struct {
	Email        string
	PasswordHash []byte
	Version      float64
}
