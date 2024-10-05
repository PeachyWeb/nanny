package handlers

import (
	"database/sql"
	"log"
	"net/http"
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
	users := []struct {
		IDuser   int
		UserName string
		Role     string
	}{}

	// Заполняем список пользователей
	for rows.Next() {
		var user struct {
			IDuser   int
			UserName string
			Role     string
		}
		err := rows.Scan(&user.IDuser, &user.UserName, &user.Role)
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
		Users    []struct {
			IDuser   int
			UserName string
			Role     string
		}
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

// Обработчик для отображения списка сотрудников
// Обработчик для отображения списка сотрудников
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
	default: // По умолчанию сортировка по ID
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
		UserName string // Убедитесь, что это имя переменной корректно
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
