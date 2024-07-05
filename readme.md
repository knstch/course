Errors

DB - 10000+
Create - 10001
Read - 10002
Update - 10003
Delete - 10004
Commit - 10010

Marshal - 10020
MarshalErr - 10021

Redis - 10030
Публикация ключа в редис - 10031
Значение по ключу не найдено в редис - 10032
Ошибка при удалении ключа из редис - 10033

Handlers - 10100
Err Reading JSON - 10101

AuthService - 11000+
Мэйл занят - 11001
Пользователь не найден - 11002
Код подтверждения неверный - 11003
Код подтверждения не найден - 11004
Не найден логин в контексте - 11005
Токен не найден - 11006
Токен устарел или невалиден - 11007
Почта пользователя уже верифицирована - 11008
Пользователь не авторизован - 11009
Юзер в бане - 11010
Юзер неактивен - 11011

UserService - 11100
Пользователь не найден - 11101
Неверно передан пароль для edit - 11102
Пароли совпадают - 11103
Старый пароль передан неверно - 11104
Фото имеет неверный формат - 11105

Ошибка минта JWT - 11010
Ошибка декодинга токена - 11011
Ошибка хеширования - 11020
Ошибка записи в formFile - 11030
Ошибка создания formFile - 11031
Ошибка подготовки запроса - 11040
Ошибка отправки запроса - 11041
Ошибка операций с байтами - 11042

Ошибка аутх в CDN - 11050
Ошибка в CDN - 11051

Контент
Ошибка при создании курса, модуля или урока, коллизиии - 13001
Модуль с таким именем не найден - 13002
Курс с таким именем не найден - 13003
Урок с таким ид не найден - 13005
Неавторизованный доступ к расширенным курсам - 13004

GRPC
14001 - ошибка при создании grpc клиента
14002 - ошибка при отправке контента по grpc

Оплата
15001 - инвойс не найден
15002 - заказ не найден
15003 - инвойс айди не совпадает с хэшем
15004 - курс уже приобретен

Адимны
16001 - логин админа занят
16002 - логин админа не найден
16003 - неверный логин или пароль
16004 - нет прав

16050 - ошибка при генерации ключа
16051 - ошибка при генерации qr кода
16052 - неверный код

Общие системные ошибки
500 

Build:
docker compose -f docker-compose-local.yml up -d --build --remove-orphans

Рабочие env
````
PORT=8080
DSN=postgres://course:password@postgres:5432/course?sslmode=disable
SECRET=ABOBA
JWT_SECRET=ABOBA
REDIS_DSN=redis://admin:password@redis:6379/0
REDIS_PASSWORD=password
REDIS_EMAIL_CHANNEL_NAME=emailKeys
ADDRESS=localhost
PG_PORT=1488
CDN_GRPC_PORT=10000
CDN_GRPC_HOST=app
ADMIN_API_KEY=aboba
CDN_API_KEY=aboba
CDN_HTTP_HOST=http://nginx:60
PG_USER=course
PG_PASSWORD=password
SUPER_ADMIN_LOGIN=admin
SUPER_ADMIN_PASSWORD=password
LOG_FILE_NAME=course
PROJECT_PATH=/home/konstantin/Desktop/course
````