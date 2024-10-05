package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

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

func CatalogPage(w http.ResponseWriter, r *http.Request) {
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

	// Получаем имя и роль пользователя из базы данных
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя из базы данных:", err)
		http.Error(w, "Ошибка при получении имени пользователя", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении роли пользователя из базы данных:", err)
		http.Error(w, "Ошибка при получении роли пользователя", http.StatusInternalServerError)
		return
	}

	// Получаем параметры фильтрации из URL-запроса
	minExperience, maxPrice, minRating, err := parseFilters(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем список нянь с учетом фильтров
	nannies, err := GetNannies(Db, minExperience, maxPrice, minRating)
	if err != nil {
		log.Println("Ошибка при получении нянь из базы данных:", err)
		http.Error(w, "Ошибка при получении нянь из базы данных", http.StatusInternalServerError)
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
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}

func NannyPage(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о текущем пользователе из сессии
	userID, userName, role, err := getUserFromSession(r)
	if err != nil {
		http.Error(w, "Не удалось получить данные пользователя из сессии", http.StatusUnauthorized)
		return
	}

	// Проверяем, передан ли ID няни
	nannyIDStr := r.URL.Query().Get("nanny_id")
	if nannyIDStr == "" {
		http.Error(w, "ID няни не передан", http.StatusBadRequest)
		return
	}

	nannyID, err := strconv.Atoi(nannyIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор няни", http.StatusBadRequest)
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
	if err := TmplNanny.Execute(w, data); err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик для страницы пользователя
func UserPage(w http.ResponseWriter, r *http.Request) {
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

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя:", err)
		http.Error(w, "Ошибка при получении имени пользователя", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении роли пользователя:", err)
		http.Error(w, "Ошибка при получении роли пользователя", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для шаблона
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
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}

func CalendarHandler(w http.ResponseWriter, r *http.Request) {

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

	// Fetch appointments from database
	query := `SELECT nannyid, starttime, endtime FROM appointments WHERE userid = $1`
	rows, err := Db.Query(query, userID)
	if err != nil {
		log.Println("Ошибка при запросе к appointments:", err)
		http.Error(w, "Unable to retrieve calendar data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.NannyID, &app.StartTime, &app.EndTime); err != nil {
			log.Println("Ошибка при сканировании записи о встрече:", err)
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		appointments = append(appointments, app)
	}

	username, err := GetUserNameByID(userID)
	if err != nil {
		http.Error(w, "Ошибка получения данных пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Calculate busy days
	busyDays := make(map[int]bool)
	for _, app := range appointments {
		day := app.StartTime.Day()
		busyDays[day] = true
	}

	// Prepare data for template
	data := struct {
		UserID   int          // Используем int вместо string
		UserName string       // Имя пользователя
		Days     []int        // Дни месяца
		BusyDays map[int]bool // Занятые дни
	}{
		UserID:   userID, // Извлекаем int userID
		UserName: username,
		Days:     make([]int, 31),
		BusyDays: busyDays,
	}

	for i := range data.Days {
		data.Days[i] = i + 1
	}

	// Execute the template
	err = TmplCalendar.Execute(w, data)
	if err != nil {
		log.Println("Ошибка при выполнении шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}
