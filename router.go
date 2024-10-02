package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"main/controllers"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter() // создаётся объект структуры или же новый маршрутизатор, mux.NewRouter() по сути конструктор
	/*Маршрутизатор связывает URL пути c конкретнвыми обрабтчиками
	может извелкать из URL переменные
	может сделать так, чтобы опреденный маршрут обрабатывал определенный метод*/

	//myController :=

	// Маршрут для регистрации
	r.HandleFunc("/", controllers.HomePage).Methods("GET")

	r.HandleFunc("/register", register).Methods("GET", "POST")

	// Маршрут для логина
	r.HandleFunc("/login", login).Methods("GET", "POST") //для аутентификации и выдачи JWT

	// Защищённые маршруты с JWT // Handle метод для связки маршрутов и обработчиков
	r.Handle("/movies", validateTokenMiddleware(http.HandlerFunc(controllers.GetMovies))).Methods("GET")
	r.Handle("/movies/create", validateTokenMiddleware(http.HandlerFunc(controllers.CreateMovie))).Methods("GET", "POST")
	r.Handle("/movies/{id}/edit", validateTokenMiddleware(http.HandlerFunc(controllers.UpdateMovie))).Methods("GET", "POST")
	r.Handle("/movies/{id}", validateTokenMiddleware(http.HandlerFunc(controllers.GetMovie))).Methods("GET")

	r.Handle("/movies/{id}/delete", validateTokenMiddleware(http.HandlerFunc(controllers.DeleteMovie))).Methods("POST")

	/*HandlerFunc позволяет использовать обычные функции с сигнатурой
	func(w http.ResponseWriter, r *http.Request) в качестве обработчиков HTTP запросов.*/

	return r
}
