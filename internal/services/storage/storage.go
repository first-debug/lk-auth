package storage

import "context"

type BlackListStorage interface {
	AddTokens(...string) error
	IsAllowed(string) (bool, error) // true если токен не в чёрном списке
	ShutDown(context.Context) error
}

type JWTStorage interface {
	AddPair(access string, refresh string) error
	GetAccessByRefresh(string) (string, error)
	ShutDown(context.Context) error
}

type UserStorage interface {
	Login(email, password string) (dataVersion float64, role string, err error)
	// Проверка на соответствие версии данных
	IsVersionValid(email string, version float64) (bool, error)
	ShutDown(context.Context) error
}
