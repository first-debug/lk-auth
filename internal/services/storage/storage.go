package storage

type BlackListStorage interface {
	AddTokens(...string) error
	IsAllowed(string) (bool, error) // true если токен не в чёрном списке
}

type JWTStorage interface {
	AddPair(access string, refresh string) error
	GetAccessByRefresh(string) (string, error)
}

type UserStorage interface {
	Login(email string, passwordHash []byte) (dataVersion float64, role string, err error)
	// Проверка на соответствие версии данных
	IsVersionValid(email string, version float64) (bool, error)
}
