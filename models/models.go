package models

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`   // Аннотирование структуры
	Username string `gorm:"unique" json:"username"` //unique для уникальности
	Password string `json:"-"`                      // "-" предотвращает отображение пароля в ответах JSON
}

type Movie struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Isbn       string    `json:"isbn"`
	Title      string    `json:"title"`
	Director   *Director `gorm:"foreignKey:DirectorID" json:"director"`
	DirectorID uint
}

type Director struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}
