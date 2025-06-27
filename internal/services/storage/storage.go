package storage

type BlackListStorage interface {
	AddTokens(...string) error
	IsAllowed(string) (bool, error) // true если токен не в чёрном списке
}

type UserStorage interface {
	Login(email, passwordHash string) (bool, error)
	// Проверка на соответствие версии данных
	IsVersionValid(email string, version float64) (bool, error)
	IncrementVersion(email string) (float64, error)
}
