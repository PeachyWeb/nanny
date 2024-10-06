package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
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

	// Проверяем роль пользователя
	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении роли пользователя:", err)
		http.Error(w, "Ошибка при получении роли пользователя", http.StatusInternalServerError)
		return
	}

	if role != "admin" {
		http.Error(w, "Недостаточно прав для доступа к странице", http.StatusForbidden)
		return
	}

	// Получаем имя пользователя
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя:", err)
		http.Error(w, "Ошибка при получении имени пользователя", http.StatusInternalServerError)
		return
	}

	// Получаем всех пользователей для отображения на странице администратора
	rows, err := Db.Query("SELECT id, login, role FROM users")
	if err != nil {
		log.Println("Ошибка при получении списка пользователей:", err)
		http.Error(w, "Ошибка при получении списка пользователей", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Определяем структуру данных для пользователей
	var users []User

	// Заполняем список пользователей
	for rows.Next() {
		var user User
		err := rows.Scan(&user.IDuser, &user.Login, &user.Role)
		if err != nil {
			log.Println("Ошибка при сканировании данных пользователя:", err)
			http.Error(w, "Ошибка при обработке данных пользователя", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Проверка на ошибки при итерации
	if err = rows.Err(); err != nil {
		log.Println("Ошибка при обработке строк пользователей:", err)
		http.Error(w, "Ошибка при обработке данных пользователей", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для шаблона
	data := struct {
		UserID   int
		UserName string
		Role     string
		Users    []User // Используем тип User
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Users:    users,
	}

	// Выполняем шаблон
	if err = TmplAdmin.Execute(w, data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return // Добавляем return после http.Error
	}
}

func AdminEmployeesPage(w http.ResponseWriter, r *http.Request) {
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

	// Проверяем, что пользователь — администратор
	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil || role != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя:", err)
		http.Error(w, "Ошибка при получении имени пользователя", http.StatusInternalServerError)
		return
	}

	// Получаем параметр сортировки из запроса
	sortBy := r.URL.Query().Get("sortBy")
	var users []User

	switch sortBy {
	case "login":
		users, err = GetAllUsersSortedByLogin(Db)
	case "role":
		users, err = GetAllUsersSortedByRole(Db)
	case "id":
		fallthrough
	default:
		users, err = GetAllUsersSortedByID(Db)
	}

	if err != nil {
		log.Println("Ошибка получения списка сотрудников:", err)
		http.Error(w, "Ошибка получения списка сотрудников", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для шаблона
	data := struct {
		UserID   int
		UserName string
		Role     string
		Users    []User
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Users:    users,
	}

	if err = TmplAdmin.Execute(w, data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}
}

// GetAllUsersSortedByID возвращает всех пользователей, отсортированных по ID
func GetAllUsersSortedByID(db *sql.DB) ([]User, error) {
	var users []User
	query := "SELECT id, login, password, role FROM users ORDER BY id"
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.IDuser, &user.Login, &user.Password, &user.Role)
		if err != nil {
			log.Println("Ошибка при сканировании данных пользователя:", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка при обработке строк пользователей:", err)
		return nil, err
	}

	return users, nil
}

// GetAllUsersSortedByLogin возвращает всех пользователей, отсортированных по логину
func GetAllUsersSortedByLogin(db *sql.DB) ([]User, error) {
	var users []User
	query := "SELECT id, login, password, role FROM users ORDER BY login"
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.IDuser, &user.Login, &user.Password, &user.Role)
		if err != nil {
			log.Println("Ошибка при сканировании данных пользователя:", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка при обработке строк пользователей:", err)
		return nil, err
	}

	return users, nil
}

// GetAllUsersSortedByRole возвращает всех пользователей, отсортированных по роли
func GetAllUsersSortedByRole(db *sql.DB) ([]User, error) {
	var users []User
	query := "SELECT id, login, password, role FROM users ORDER BY role"
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.IDuser, &user.Login, &user.Password, &user.Role)
		if err != nil {
			log.Println("Ошибка при сканировании данных пользователя:", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка при обработке строк пользователей:", err)
		return nil, err
	}

	return users, nil
}

func getUserFromSession(r *http.Request) (int, string, string, error) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return 0, "", "", err
	}

	// Проверяем, если в сессии нет данных о пользователе
	userID, ok := session.Values["userID"].(int)
	if !ok {
		return 0, "", "", http.ErrNoCookie
	}

	userName, ok := session.Values["userName"].(string)
	if !ok {
		return 0, "", "", http.ErrNoCookie
	}

	role, ok := session.Values["role"].(string)
	if !ok {
		return 0, "", "", http.ErrNoCookie
	}

	return userID, userName, role, nil
}

// parseFilters получает параметры фильтрации из URL-запроса и возвращает их в виде числовых значений.
func parseFilters(r *http.Request) (int, int, int, error) {
	// Получаем параметры фильтрации из URL-запроса
	minExperienceStr := r.URL.Query().Get("min_experience")
	maxPriceStr := r.URL.Query().Get("max_price")
	minRatingStr := r.URL.Query().Get("min_rating")

	var err error
	var minExperience, maxPrice, minRating int

	// Преобразуем параметры в числовые значения, если они переданы
	if minExperienceStr != "" {
		minExperience, err = strconv.Atoi(minExperienceStr)
		if err != nil {
			return 0, 0, 0, errors.New("Неверный формат параметра min_experience")
		}
	}

	if maxPriceStr != "" {
		maxPrice, err = strconv.Atoi(maxPriceStr)
		if err != nil {
			return 0, 0, 0, errors.New("Неверный формат параметра max_price")
		}
	}

	if minRatingStr != "" {
		minRating, err = strconv.Atoi(minRatingStr)
		if err != nil {
			return 0, 0, 0, errors.New("Неверный формат параметра min_rating")
		}
	}

	// Возвращаем значения параметров фильтрации
	return minExperience, maxPrice, minRating, nil
}
