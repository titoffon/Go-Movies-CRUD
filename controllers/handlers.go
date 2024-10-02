package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"main/db" // Импортируем пакет db
	"main/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html", "templates/base.html")
	if err != nil {
		log.Printf("Ошибка при парсинге шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base.html", nil); err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func GetMovies(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос на получение всех фильмов")
	username := r.Context().Value("username").(string)

	var movies []models.Movie
	if err := db.DB.Preload("Director").Find(&movies).Error; err != nil {
		log.Printf("Ошибка при получении фильмов из базы данных: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/movies.html", "templates/base.html")
	if err != nil {
		log.Printf("Ошибка при парсинге шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
		Movies   []models.Movie
		MyTitle  string
	}{
		Username: username,
		Movies:   movies,
		MyTitle:  "Список фильмов",
	}

	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handlers.go
func GetMovie(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос на получение информации о фильме")

	// Получаем имя пользователя из контекста
	usernameInterface := r.Context().Value("username")
	if usernameInterface == nil {
		log.Println("getMovie: Не удалось получить имя пользователя из контекста")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := usernameInterface.(string)
	if !ok {
		log.Println("getMovie: Имя пользователя в контексте имеет неверный тип")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Получаем ID фильма из URL
	params := mux.Vars(r)
	movieIDStr := params["id"]

	// Преобразуем строку в uint
	movieIDUint64, err := strconv.ParseUint(movieIDStr, 10, 64)
	if err != nil {
		log.Printf("Недопустимый формат ID фильма: %v", err)
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}
	movieID := uint(movieIDUint64)

	// Получаем фильм из базы данных
	var movie models.Movie
	if err := db.DB.Preload("Director").First(&movie, movieID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Фильм с ID %d не найден", movieID)
			http.Error(w, "Movie not found", http.StatusNotFound)
		} else {
			log.Printf("Ошибка при получении фильма: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Отображаем страницу с информацией о фильме
	tmpl, err := template.ParseFiles("templates/movie.html", "templates/base.html")
	if err != nil {
		log.Printf("Ошибка при парсинге шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
		Movie    models.Movie
	}{
		Username: username,
		Movie:    movie,
	}

	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func CreateMovie(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос на создание нового фильма")

	usernameInterface := r.Context().Value("username")
	if usernameInterface == nil {
		log.Println("createMovie: Не удалось получить имя пользователя из контекста")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := usernameInterface.(string)
	if !ok {
		log.Println("createMovie: Имя пользователя в контексте имеет неверный тип")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		// Отображаем форму для создания фильма
		tmpl, err := template.ParseFiles("templates/movie_new.html", "templates/base.html")
		if err != nil {
			log.Printf("Ошибка при парсинге шаблона: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := struct {
			Username          string
			Error             string
			Title             string
			Isbn              string
			DirectorFirstname string
			DirectorLastname  string
		}{
			Username:          username,
			Error:             "",
			Title:             "",
			Isbn:              "",
			DirectorFirstname: "",
			DirectorLastname:  "",
		}

		if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
			log.Printf("Ошибка при выполнении шаблона: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Обработка данных формы
		title := r.FormValue("title")
		isbn := r.FormValue("isbn")
		directorFirstname := r.FormValue("director_firstname")
		directorLastname := r.FormValue("director_lastname")

		// Проверка введённых данных
		if title == "" || isbn == "" || directorFirstname == "" || directorLastname == "" {
			tmpl, err := template.ParseFiles("templates/movie_new.html", "templates/base.html")
			if err != nil {
				log.Printf("Ошибка при парсинге шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			data := struct {
				Username          string
				Error             string
				Title             string
				Isbn              string
				DirectorFirstname string
				DirectorLastname  string
			}{
				Username:          username,
				Error:             "Все поля обязательны для заполнения.",
				Title:             title,
				Isbn:              isbn,
				DirectorFirstname: directorFirstname,
				DirectorLastname:  directorLastname,
			}
			if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
				log.Printf("Ошибка при выполнении шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Проверяем, существует ли режиссёр с указанным именем и фамилией
		var director models.Director
		result := db.DB.First(&director, "firstname = ? AND lastname = ?", directorFirstname, directorLastname)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Режиссёр не найден, создаём нового
				director = models.Director{
					Firstname: directorFirstname,
					Lastname:  directorLastname,
				}
				if err := db.DB.Create(&director).Error; err != nil {
					log.Printf("Ошибка при создании режиссёра: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else {
				// Произошла другая ошибка при обращении к базе данных
				log.Printf("Ошибка при поиске режиссёра: %v", result.Error)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Создание записи фильма с использованием найденного или созданного режиссёра
		movie := models.Movie{
			Isbn:       isbn,
			Title:      title,
			DirectorID: director.ID,
		}
		if err := db.DB.Create(&movie).Error; err != nil {
			log.Printf("Ошибка при создании фильма: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Перенаправление на страницу со списком фильмов
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
		return
	}

	// Если метод не GET и не POST, возвращаем ошибку 405
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handlers.go
func UpdateMovie(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос на обновление фильма")

	// Получаем имя пользователя из контекста
	usernameInterface := r.Context().Value("username")
	if usernameInterface == nil {
		log.Println("updateMovie: Не удалось получить имя пользователя из контекста")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := usernameInterface.(string)
	if !ok {
		log.Println("updateMovie: Имя пользователя в контексте имеет неверный тип")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Получаем ID фильма из URL
	params := mux.Vars(r)
	movieIDStr := params["id"]

	// Преобразуем строку в uint
	movieIDUint64, err := strconv.ParseUint(movieIDStr, 10, 64)
	if err != nil {
		log.Printf("Недопустимый формат ID фильма: %v", err)
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}
	movieID := uint(movieIDUint64)

	if r.Method == http.MethodGet {
		// Получаем фильм из базы данных
		var movie models.Movie
		if err := db.DB.Preload("Director").First(&movie, movieID).Error; err != nil {
			log.Printf("Фильм с ID %d не найден: %v", movieID, err)
			http.Error(w, "Movie not found", http.StatusNotFound)
			return
		}

		// Отображаем форму для редактирования фильма
		tmpl, err := template.ParseFiles("templates/movie_edit.html", "templates/base.html")
		if err != nil {
			log.Printf("Ошибка при парсинге шаблона: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := struct {
			Username          string
			Movie             models.Movie
			DirectorFirstname string
			DirectorLastname  string
			Error             string
		}{
			Username:          username,
			Movie:             movie,
			DirectorFirstname: movie.Director.Firstname,
			DirectorLastname:  movie.Director.Lastname,
			Error:             "",
		}

		if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
			log.Printf("Ошибка при выполнении шаблона: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Обработка данных формы
		title := r.FormValue("title")
		isbn := r.FormValue("isbn")
		directorFirstname := r.FormValue("director_firstname")
		directorLastname := r.FormValue("director_lastname")

		// Проверка введённых данных
		if title == "" || isbn == "" || directorFirstname == "" || directorLastname == "" {
			tmpl, err := template.ParseFiles("templates/movie_edit.html", "templates/base.html")
			if err != nil {
				log.Printf("Ошибка при парсинге шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			data := struct {
				Username          string
				Movie             models.Movie
				DirectorFirstname string
				DirectorLastname  string
				Error             string
			}{
				Username:          username,
				Movie:             models.Movie{ID: movieID, Title: title, Isbn: isbn},
				DirectorFirstname: directorFirstname,
				DirectorLastname:  directorLastname,
				Error:             "Все поля обязательны для заполнения.",
			}

			if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
				log.Printf("Ошибка при выполнении шаблона: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Начинаем транзакцию для обновления фильма и режиссёра
		tx := db.DB.Begin()

		// Обновляем или создаём режиссёра
		var director models.Director
		result := db.DB.First(&director, "firstname = ? AND lastname = ?", directorFirstname, directorLastname)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Режиссёр не найден, создаём нового
				director = models.Director{
					Firstname: directorFirstname,
					Lastname:  directorLastname,
				}
				if err := tx.Create(&director).Error; err != nil {
					tx.Rollback()
					log.Printf("Ошибка при создании режиссёра: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else {
				// Произошла другая ошибка при обращении к базе данных
				tx.Rollback()
				log.Printf("Ошибка при поиске режиссёра: %v", result.Error)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Обновляем фильм
		updates := models.Movie{
			Title:      title,
			Isbn:       isbn,
			DirectorID: director.ID,
		}
		if err := tx.Model(&models.Movie{}).Where("id = ?", movieID).Updates(updates).Error; err != nil {
			tx.Rollback()
			log.Printf("Ошибка при обновлении фильма: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		tx.Commit()

		// Перенаправление на страницу с подробной информацией о фильме
		http.Redirect(w, r, fmt.Sprintf("/movies/%d", movieID), http.StatusSeeOther)
		return
	}

	// Если метод не GET и не POST, возвращаем ошибку 405
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handlers.go
func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос на удаление фильма")

	// Получаем имя пользователя из контекста
	usernameInterface := r.Context().Value("username")
	if usernameInterface == nil {
		log.Println("deleteMovie: Не удалось получить имя пользователя из контекста")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, ok := usernameInterface.(string)
	if !ok {
		log.Println("deleteMovie: Имя пользователя в контексте имеет неверный тип")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Получаем ID фильма из URL
	params := mux.Vars(r)
	movieIDStr := params["id"]

	// Преобразуем строку в uint
	movieIDUint64, err := strconv.ParseUint(movieIDStr, 10, 64)
	if err != nil {
		log.Printf("Недопустимый формат ID фильма: %v", err)
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}
	movieID := uint(movieIDUint64)

	// Находим фильм
	var movie models.Movie
	if err := db.DB.First(&movie, movieID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Фильм с ID %d не найден", movieID)
			http.Error(w, "Movie not found", http.StatusNotFound)
		} else {
			log.Printf("Ошибка при получении фильма: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Удаляем фильм
	if err := db.DB.Delete(&movie).Error; err != nil {
		log.Printf("Ошибка при удалении фильма: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу со списком фильмов
	http.Redirect(w, r, "/movies", http.StatusSeeOther)
}
