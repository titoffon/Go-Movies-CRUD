package main

import (
	"fmt"
	"log"
	"main/db"

	//"main/models"
	"net/http"
)

//var movies []models.Movie

func main() {
	// Инициализация базы данных
	db.InitDB()

	// Настройка маршрутизатора
	r := SetupRouter()

	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Запуск сервера
	fmt.Printf("Starting server at port 8000\n")
	log.Fatal(http.ListenAndServe(":8000", r))
}
