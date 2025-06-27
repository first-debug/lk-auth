// В пакете libs лежат вспомогательные типы и функции.
// В этом файле реализован удобный вариант логирования ошибок.
package sl

import "log/slog"

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
