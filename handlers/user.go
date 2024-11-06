package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

// ProfilePage отображает страницу профиля пользователя
func ProfilePage(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Println("Ошибка при получении сессии:", err)
		http.Error(w, "Ошибка при получении сессии", http.StatusInternalServerError)
		return
	}

	// Проверяем идентификатор пользователя в сессии
	userID, ok := session.Values["userID"].(int)
	if !ok || userID <= 0 {
		http.Error(w, "Необходимо войти в систему", http.StatusUnauthorized)
		return
	}

	// Получаем информацию о пользователе
	user, err := GetUserByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении информации о пользователе:", err)
		http.Error(w, "Ошибка при получении информации о пользователе", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для шаблона
	data := struct {
		UserID     int
		UserName   string
		FirstName  string
		LastName   string
		Patronymic string
		City       string
		Phone      string
		Age        int64
	}{
		UserID:     user.IDuser,
		UserName:   user.Login,
		FirstName:  user.FirstName.String,
		LastName:   user.LastName.String,
		Patronymic: user.Patronymic.String,
		City:       user.City.String,
		Phone:      user.Phone.String,
		Age:        user.Age.Int64,
	}

	// Выполняем шаблон
	if err = TmplProfile.Execute(w, data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}
}

// GetUserByIDFromDB возвращает информацию о пользователе по его ID
func GetUserByIDFromDB(userID int) (User, error) {
	var user User
	query := "SELECT id, login, first_name, last_name, patronymic, city, phone, age FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&user.IDuser, &user.Login, &user.FirstName, &user.LastName, &user.Patronymic, &user.City, &user.Phone, &user.Age)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return user, err
	}
	return user, nil
}

// UpdateProfile обрабатывает изменения профиля пользователя
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем сессию
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Println("Ошибка при получении сессии:", err)
		http.Error(w, "Ошибка при получении сессии", http.StatusInternalServerError)
		return
	}

	// Проверяем идентификатор пользователя в сессии
	userID, ok := session.Values["userID"].(int)
	if !ok || userID <= 0 {
		http.Error(w, "Необходимо войти в систему", http.StatusUnauthorized)
		return
	}

	// Получаем новые данные из формы
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	patronymic := r.FormValue("patronymic")
	city := r.FormValue("city")
	phone := r.FormValue("phone")
	ageStr := r.FormValue("age")

	var age sql.NullInt64
	if ageStr != "" {
		parsedAge, err := strconv.ParseInt(ageStr, 10, 64)
		if err != nil {
			log.Println("Ошибка при преобразовании возраста:", err)
			http.Error(w, "Некорректный возраст", http.StatusBadRequest)
			return
		}
		age = sql.NullInt64{Int64: parsedAge, Valid: true}
	} else {
		age = sql.NullInt64{Valid: false} // Устанавливаем значение NULL
	}

	// Обновляем данные пользователя в базе данных
	query := `
        UPDATE users 
        SET 
            first_name = $1, 
            last_name = $2, 
            patronymic = $3, 
            city = $4, 
            phone = $5, 
            age = $6 
        WHERE id = $7`
	_, err = Db.Exec(query, firstName, lastName, patronymic, city, phone, age, userID)
	if err != nil {
		log.Println("Ошибка при обновлении данных пользователя:", err)
		http.Error(w, "Ошибка при обновлении данных пользователя", http.StatusInternalServerError)
		return
	}

	// Перенаправляем пользователя обратно на страницу профиля
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
