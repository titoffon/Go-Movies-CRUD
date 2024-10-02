package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"main/db"
	"main/models"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtKey = []byte("my_secret_key") // Секретный ключ для подписи JWT

/*// Структура для хранения данных пользователя (обычно берется из базы данных)
var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}*/

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

/*func register(w http.ResponseWriter, r *http.Request) {
var creds Credentials
err := json.NewDecoder(r.Body).Decode(&creds) // r объект типа *http.Request, представляющий входящий http запрос
/*r.Body содержит в себе тело запроса для будущего чтения данных
json.NewDecoder(r.Body) создание декодера(объект, который будет читать данные запросы
Decode(&creds) метод декодера, который декодирует json в структуру*/
/*if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь
	var existingUser models.User
	result := db.DB.First(&existingUser, "username = ?", creds.Username) // запрос к БД с поискоим первого совпадения по юзернейму
	if result.Error == nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создаем нового пользователя
	user := models.User{
		Username: creds.Username,
		Password: string(hashedPassword),
	}
	if err := db.DB.Create(&user).Error; err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated) //Устанавливается статус ответа 201 Created, указывающий на успешное создание ресурса
}*/

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/register.html", "templates/base.html")
		if err != nil {
			log.Printf("Ошибка при парсинге шаблона: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base.html", nil); err != nil {
			log.Printf("Ошибка при выполнении шаблона: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Обработка данных формы
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Проверка, что имя пользователя и пароль не пустые
		if username == "" || password == "" {
			tmpl, err := template.ParseFiles("templates/register.html", "templates/base.html")
			if err != nil {
				log.Printf("Ошибка при парсинге шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			data := struct {
				Error string
			}{
				Error: "Имя пользователя и пароль обязательны для заполнения.",
			}
			if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
				log.Printf("Ошибка при выполнении шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Проверяем, существует ли пользователь с таким именем
		var existingUser models.User
		result := db.DB.First(&existingUser, "username = ?", username)
		if result.Error == nil {
			// Пользователь уже существует
			tmpl, err := template.ParseFiles("templates/register.html", "templates/base.html")
			if err != nil {
				log.Printf("Ошибка при парсинге шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			data := struct {
				Error string
			}{
				Error: "Пользователь с таким именем уже существует.",
			}
			if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
				log.Printf("Ошибка при выполнении шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		} else if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Произошла другая ошибка при обращении к базе данных
			log.Printf("Ошибка при проверке пользователя: %v", result.Error)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Хэшируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Ошибка при хэшировании пароля: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Создаем нового пользователя
		user := models.User{
			Username: username,
			Password: string(hashedPassword),
		}
		if err := db.DB.Create(&user).Error; err != nil {
			log.Printf("Ошибка при создании пользователя: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу логина после успешной регистрации
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

}

// Хендлер для аутентификации пользователя
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render the login page
		tmpl, err := template.ParseFiles("templates/login.html", "templates/base.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base.html", nil); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Process form data
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Validate input
		if username == "" || password == "" {
			tmpl, err := template.ParseFiles("templates/login.html", "templates/base.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			data := struct {
				Error string
			}{
				Error: "Имя пользователя и пароль обязательны для заполнения.",
			}
			if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Authenticate user
		var user models.User
		result := db.DB.First(&user, "username = ?", username)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// User not found
				tmpl, err := template.ParseFiles("templates/login.html", "templates/base.html")
				if err != nil {
					log.Printf("Error parsing template: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				data := struct {
					Error string
				}{
					Error: "Неверное имя пользователя или пароль.",
				}
				if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
					log.Printf("Error executing template: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			// Other database error
			log.Printf("Error retrieving user: %v", result.Error)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Compare password
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			// Invalid password
			tmpl, err := template.ParseFiles("templates/login.html", "templates/base.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			data := struct {
				Error string
			}{
				Error: "Неверное имя пользователя или пароль.",
			}
			if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Create JWT token
		expirationTime := time.Now().Add(5 * time.Minute)
		claims := &Claims{
			Username: username,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			log.Printf("Error creating JWT token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the token in a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
			Path:     "/",
		})

		// Redirect to the protected page
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
		return
	}

	// If method is not GET or POST, return 405 Method Not Allowed
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func validateTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("validateTokenMiddleware: Проверка JWT токена")

		// Получаем токен из заголовков или куки
		cookie, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				log.Println("validateTokenMiddleware: Токен не найден в куки")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Printf("validateTokenMiddleware: Ошибка при чтении куки: %v\n", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Получаем токен строку
		tokenString := cookie.Value
		log.Printf("validateTokenMiddleware: Токен получен: %s\n", tokenString)

		// Проверяем токен
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				log.Println("validateTokenMiddleware: Подпись JWT токена недействительна")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Printf("validateTokenMiddleware: Ошибка при парсинге токена: %v\n", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if !token.Valid {
			log.Println("validateTokenMiddleware: Токен недействителен")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("validateTokenMiddleware: Токен валиден для пользователя: %s\n", claims.Username)

		// Устанавливаем пользователя в контекст запроса, если токен валиден
		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
