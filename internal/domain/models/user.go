// Доменная область
package models

type User struct {
	Email        string `redis:"email"`
	PasswordHash string `redis:"passHash"`
	// TODO: добавить валидацию для поля Role
	// Пример: `validate:"required,oneof=admin user guest"`
	Role    string  `redis:"role"`
	Version float64 `redis:"version"`
}
