package tests

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/knstch/course/internal/app"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/router"
	"github.com/stretchr/testify/assert"
)

type tests struct {
	name    string
	want    want
	request request
}

type want struct {
	statusCode int
	body       string
}

type request struct {
	contentType string
	body        string
}

type testUser struct {
	email    string
	password string
	cookie   []*http.Cookie
}

const (
	env = `
		PORT=8080
		DSN=postgres://course:password@localhost:1488/course?sslmode=disable
		SECRET=ABOBA
		JWT_SECRET=ABOBA
		REDIS_DSN=redis://admin:password@localhost:6379/0
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
		SERVICE_EMAIL=kostyacherepanov1@gmail.com
		SERVICE_EMAIL_PASSWORD="rwsw qefe tdxk fgxl"
		SMPT_HOST=smtp.gmail.com
		SMPT_PORT=587
		IS_TEST=true
	`

	newPassForUserOne = "Passs@0101"
)

var (
	userMain = testUser{
		email:    "konstchere@gmail.com",
		password: "Xer@0101",
	}

	userOne = testUser{
		email:    fmt.Sprintf("%s@gmail.com", randomString(7)),
		password: "Xer@0101",
	}

	testsConfig *config.Config

	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func randomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestRegistration(t *testing.T) {
	tests := []tests{
		{
			name: "#1 нормальная регистрация",
			want: want{
				statusCode: http.StatusOK,
				body: `{
    				"message": "пользователь зарегистрирован",
    				"success": true
					}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "%s","password": "%s"}`, userOne.email, userOne.password),
			},
		},
		{
			name: "#2 пользователь уже зарегистрирован",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    						"error": "пользователь с таким email уже существует",
    						"code": 11001
						}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "%s","password": "%s"}`, userMain.email, userMain.password),
			},
		},
		{
			name: "#3 указан неверный email",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    						"error": "email: email передан неправильно.",
    						"code": 400
						}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "%s","password": "%s"}`, "asdasd.ru", userMain.password),
			},
		},
		{
			name: "#4 указан неверный пароль, меньше 8 символов",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    						"error": "password: пароль должен содержать как миниум 8 символов.",
    						"code": 400
						}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "%s","password": "%s"}`, userMain.email, "Xer@011"),
			},
		},
		{
			name: "#5 указан неверный пароль, без upper case",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    						"error": "password: пароль должен содержать как минимум 1 букву, 1 заглавную букву и 1 цифру.",
    						"code": 400
						}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "%s","password": "%s"}`, userMain.email, "xer@0101"),
			},
		},
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Print(err)
		return
	}

	envFilePath := filepath.Join(dir, ".env")

	file, err := os.Create(envFilePath)
	if err != nil {
		log.Print(err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(env)
	if err != nil {
		log.Print(err)
		return
	}

	if err := config.InitENV(dir); err != nil {
		log.Print(err)
		return
	}

	testsConfig = config.GetConfig()

	container, errs := app.InitContainer(dir, testsConfig)
	if errs != nil {
		log.Print(errs)
		return
	}

	router := router.RequestsRouter(container.Handlers)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/auth/register", bytes.NewBuffer([]byte(tt.request.body)))
			req.Header.Set("Content-Type", tt.request.contentType)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Print(err)
				return
			}

			assert.Equal(t, tt.want.statusCode, resp.Code)
			assert.JSONEq(t, tt.want.body, string(body))

			if tt.name == "#1 нормальная регистрация" {
				userOne.cookie = append(userOne.cookie, resp.Result().Cookies()...)
			}
		})
	}
}

func TestVerification(t *testing.T) {
	tests := []tests{
		{
			name: "#1 код верификации невалидный",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    				"error": "Code: код верификации передан неверно.",
    				"code": 400
					}`,
			},
			request: request{
				body: "?confirmCode=111",
			},
		},
		{
			name: "#2 код верификации неверный",
			want: want{
				statusCode: http.StatusForbidden,
				body: `{
    				"error": "код подтверждения не правильный",
    				"code": 11003
					}`,
			},
			request: request{
				body: "?confirmCode=2222",
			},
		},
		{
			name: "#3 нет куки",
			want: want{
				statusCode: http.StatusForbidden,
				body: `{
    				"error": "пользователь не авторизован",
    				"code": 11009
					}`,
			},
			request: request{
				body: "?confirmCode=2222",
			},
		},
		{
			name: "#4 нормальная верификация",
			want: want{
				statusCode: http.StatusOK,
				body: `{
    				"message": "email верифицирован",
    				"success": true
					}`,
			},
			request: request{
				body: "?confirmCode=1111",
			},
		},
		{
			name: "#5 email уже верифицирован",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    				"error": "почта пользователя уже верифицирована",
    				"code": 11008
					}`,
			},
			request: request{
				body: "?confirmCode=1111",
			},
		},
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Print(err)
		return
	}

	container, err := app.InitContainer(dir, testsConfig)
	if err != nil {
		log.Print(err)
		return
	}

	router := router.RequestsRouter(container.Handlers)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/v1/auth/email/verification%s", tt.request.body), nil)
			if tt.name != "#3 нет куки" {
				for _, v := range userOne.cookie {
					req.AddCookie(v)
				}
			}

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Print(err)
				return
			}

			assert.Equal(t, tt.want.statusCode, resp.Code)
			assert.JSONEq(t, tt.want.body, string(body))

			if tt.name == "#4 нормальная верификация" {
				userOne.cookie = userOne.cookie[:0]
				userOne.cookie = append(userOne.cookie, resp.Result().Cookies()...)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []tests{
		{
			name: "#1 нормальный логин",
			want: want{
				statusCode: http.StatusOK,
				body: `{
    				"message": "доступ разрешен",
    				"success": true
					}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "%s","password": "%s"}`, userOne.email, userOne.password),
			},
		},
		{
			name: "#2 несуществующий пользователь",
			want: want{
				statusCode: http.StatusNotFound,
				body: `{
    				"error": "пользователь не найден",
    				"code": 11002
					}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "aboba-1234@gmail.com","password": "%s"}`, userOne.password),
			},
		},
		{
			name: "#3 не передан email",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    				"error": "email: email обязательно.",
    				"code": 400
					}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "","password": "%s"}`, userOne.password),
			},
		},
		{
			name: "#4 не передан пароль",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
    				"error": "password: пароль обязателен.",
    				"code": 400
					}`,
			},
			request: request{
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"email": "%s","password": ""}`, userOne.email),
			},
		},
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Print(err)
		return
	}

	container, err := app.InitContainer(dir, testsConfig)
	if err != nil {
		log.Print(err)
		return
	}

	router := router.RequestsRouter(container.Handlers)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/auth/login", bytes.NewBuffer([]byte(tt.request.body)))

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Print(err)
				return
			}

			assert.Equal(t, tt.want.statusCode, resp.Code)
			assert.JSONEq(t, tt.want.body, string(body))
		})
	}
}

func TestSendRecoverPasswordCode(t *testing.T) {
	tests := []struct {
		name    string
		want    want
		request request
	}{
		{
			name: "#1 Невалидный email",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
					"error": "email: email передан неправильно.",
					"code": 400
				}`,
			},
			request: request{
				body: "?email=aboba",
			},
		},
		{
			name: "#2 успешная отправка",
			want: want{
				statusCode: http.StatusOK,
				body: `{
					"message": "код для восстановления успешно отправлен",
					"success": true
				}`,
			},
			request: request{
				body: fmt.Sprintf("?email=%s", userOne.email),
			},
		},
		{
			name: "#3 Слишком много запросов",
			want: want{
				statusCode: http.StatusTooManyRequests,
				body: `{
					"error": "письмо уже было отправлено, подождите 1 минуту перед отправкой нового",
					"code": 17002
				}`,
			},
			request: request{
				body: fmt.Sprintf("?email=%s", userOne.email),
			},
		},
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Print(err)
		return
	}

	container, err := app.InitContainer(dir, testsConfig)
	if err != nil {
		log.Print(err)
		return
	}

	router := router.RequestsRouter(container.Handlers)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/api/v1/auth/sendRecoveryCode%s", tt.request.body), nil)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Print(err)
				return
			}

			assert.Equal(t, tt.want.statusCode, resp.Code)
			assert.JSONEq(t, tt.want.body, string(body))
		})
	}
}

func TestSetNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		want    want
		request request
	}{
		{
			name: "#1 невалидный JSON",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
					"error": "запрос передан в неверном формате",
					"code": 10101
				}`,
			},
			request: request{
				body: `{"email": "notanemail", "password": "newpassword", "code": "1234"`,
			},
		},
		{
			name: "#2 неверный код подтверждения",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
					"error": "код подтверждения неверный",
					"code": 11003
				}`,
			},
			request: request{
				body: fmt.Sprintf(`{"email": "%s", "password": "%s", "code": 2222}`, userOne.email, newPassForUserOne),
			},
		},
		{
			name: "#3 невалидный код подтверждения",
			want: want{
				statusCode: http.StatusBadRequest,
				body: `{
					"error": "код верификации передан неверно",
					"code": 400
				}`,
			},
			request: request{
				body: fmt.Sprintf(`{"email": "%s", "password": "%s", "code": 11111}`, userOne.email, newPassForUserOne),
			},
		},
		{
			name: "#4 Успешное восстановление пароля",
			want: want{
				statusCode: http.StatusOK,
				body: `{
					"message": "пароль успешно восстановлен",
					"success": true
				}`,
			},
			request: request{
				body: fmt.Sprintf(`{"email": "%s", "password": "%s", "code": "1111"}`, userOne.email, newPassForUserOne),
			},
		},
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Print(err)
		return
	}

	container, err := app.InitContainer(dir, testsConfig)
	if err != nil {
		log.Print(err)
		return
	}

	router := router.RequestsRouter(container.Handlers)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/auth/recoverPassword", bytes.NewReader([]byte(tt.request.body)))
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Print(err)
				return
			}

			assert.Equal(t, tt.want.statusCode, resp.Code)
			assert.JSONEq(t, tt.want.body, string(body))
		})
	}
}
