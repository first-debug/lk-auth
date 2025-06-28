// Доменная область
package models

type User struct {
	Email        string
	PasswordHash []byte
	Version      float64
	// TODO: добавить валидацию для поля Role
	// Пример: `validate:"required,oneof=admin user guest"`
	Role string
}
