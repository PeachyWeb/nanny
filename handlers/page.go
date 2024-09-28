package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

// Обработчик для страницы администратора
func AdminPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	err = TmplAdmin.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик для страницы пользователя
func UserPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	err = TmplMain.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик для отображения списка сотрудников
func AdminEmployeesPage(w http.ResponseWriter, r *http.Request) {
	// Получаем список всех пользователей
	users, err := GetAllUsers()
	if err != nil {
		http.Error(w, "Ошибка получения списка сотрудников: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Предположим, что текущий пользователь — это администратор, информация о котором нам также нужна
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Получаем информацию о текущем пользователе
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка получения имени пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка получения роли пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Подготовка данных для передачи в шаблон
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

	// Выполняем рендеринг шаблона с обновленной структурой данных
	err = TmplAdmin.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}
func HomePage(w http.ResponseWriter, r *http.Request) {
	// Получаем текущего пользователя (это нужно сделать после аутентификации)
	var currentUser User
	// Здесь вы должны получить информацию о текущем пользователе, например, из сессии или токена
	// currentUser = GetCurrentUser(r) // Реализуйте получение текущего пользователя

	// Передаем информацию о пользователе в шаблон
	tmpl := ParseTemplate("templates/home.html")
	tmpl.Execute(w, currentUser) // Передаем текущего пользователя в шаблон
}

func CatalogPage(w http.ResponseWriter, r *http.Request) {
	// Получаем текущего пользователя через запрос (например, ID пользователя)
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Получаем имя и роль пользователя из базы данных
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Получаем список нянь из базы данных
	nannies, err := GetNannies(Db) // Предположим, что у вас есть переменная db, которая является *sql.DB
	if err != nil {
		http.Error(w, "Ошибка при получении нянь из базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Заполняем данные для страницы
	data := PageData{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Nannies:  nannies,
	}

	// Рендерим шаблон
	tmpl := ParseTemplate("templates/upload1.html")
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func NannyPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	nannyIDStr := r.URL.Query().Get("nanny_id")

	// Проверяем, передан ли ID пользователя
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Проверяем, передан ли ID няни
	if nannyIDStr == "" {
		http.Error(w, "ID няни не передан", http.StatusBadRequest)
		return
	}

	nannyID, err := strconv.Atoi(nannyIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор няни", http.StatusBadRequest)
		return
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Получаем информацию о конкретной няне из базы данных
	var nanny Nanny
	err = Db.QueryRow("SELECT id, name, description, price, photo_url FROM nannies WHERE id = $1", nannyID).Scan(&nanny.ID, &nanny.Name, &nanny.Description, &nanny.Price, &nanny.PhotoURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Няня не найдена", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка при получении информации о няне", http.StatusInternalServerError)
		}
		return
	}

	// Получаем отзывы о няне и вычисляем средний рейтинг
	reviews, err := GetReviewsByNannyID(nannyID)
	if err != nil {
		http.Error(w, "Ошибка при получении отзывов из базы данных", http.StatusInternalServerError)
		return
	}

	// Вычисление среднего рейтинга
	var totalRating float64
	for _, review := range reviews {
		totalRating += float64(review.Rating)
	}
	averageRating := 0.0
	if len(reviews) > 0 {
		averageRating = totalRating / float64(len(reviews))
	}

	// Обновляем средний рейтинг в базе данных
	err = UpdateNannyRating(nannyID)
	if err != nil {
		log.Println("Ошибка при обновлении рейтинга няни:", err)
	}

	// Подготовка данных для передачи в шаблон
	data := NannyDetailPage{
		UserID:        userID,
		UserName:      userName,
		Role:          role,
		Nanny:         nanny,
		AverageRating: averageRating,
		Reviews:       reviews,
	}

	// Рендерим шаблон
	err = TmplNanny.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}
