package cdnerrors

import "errors"

var (
	ErrBadFile          = errors.New("загруженный файл имеет неверный формат")
	ErrFailedAuth       = errors.New("ошибка авторизации, неверный API ключ или отсутствует userId")
	ErrCdnFailture      = errors.New("ошибка в CDN")
	ErrCdnNotResponding = errors.New("CDN не отвечает")
)
