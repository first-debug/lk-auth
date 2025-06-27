// Здесь должна быть бизнес логика ответсвенная за авторизацию
package auth

type AuthService interface {
	Login(email, password string) (string, string, error)
	Refresh(string) (string, string, error)
	ValidateToken(string) (bool, error)
	Logout(...string) error
}
