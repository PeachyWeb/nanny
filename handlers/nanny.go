package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

// Получаем рейтинги для конкретной няни
func GetRatingsForNanny(nannyID int) ([]float64, error) {
	var ratings []float64
	rows, err := Db.Query("SELECT rating FROM ratings WHERE nanny_id = ?", nannyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rating float64
		if err := rows.Scan(&rating); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}
	return ratings, nil
}
func GetNannies(db *sql.DB, minExperience int, maxPrice int, minRating int) ([]Nanny, error) {
	var nannies []Nanny

	// Базовый SQL-запрос
	query := `SELECT id, name, experience, phone, description, price, photo_url, average_rating, review_count, city FROM nannies WHERE 1=1`

	// Добавляем условия фильтрации, если они заданы
	var args []interface{}
	if minExperience > 0 {
		query += ` AND experience >= ?`
		args = append(args, minExperience)
	}
	if maxPrice > 0 {
		query += ` AND price <= ?`
		args = append(args, maxPrice)
	}
	if minRating > 0 {
		query += ` AND average_rating >= ?`
		args = append(args, minRating)
	}

	// Выполняем запрос с фильтрами
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса: %v", err)
		return nil, err
	}
	defer rows.Close() // Закрываем rows после использования

	// Проходим по результатам запроса
	for rows.Next() {
		var nanny Nanny
		// Сканируем данные из строки в структуру
		if err := rows.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.AverageRating, &nanny.ReviewCount, &nanny.City); err != nil {
			log.Printf("Ошибка при сканировании строки: %v", err)
			continue // Пропускаем ошибочную строку
		}
		nannies = append(nannies, nanny) // Добавляем няню в срез
	}

	// Проверяем на ошибки, возникшие во время итерации по строкам
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка в процессе чтения строк: %v", err)
		return nil, err
	}

	return nannies, nil // Возвращаем полученный список нянь
}

// NannyHandler отображает страницу няни
func NannyHandler(w http.ResponseWriter, r *http.Request) {
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

	// Получаем идентификатор няни из URL
	idNanny := r.URL.Query().Get("nanny_id")
	if idNanny == "" {
		http.Error(w, "Отсутствует ID няни", http.StatusBadRequest)
		return
	}

	// Получаем информацию о няне из базы данных
	query := "SELECT id, name, experience, phone, average_rating, description, price, photo_url, city FROM nannies WHERE id = $1"
	row := Db.QueryRow(query, idNanny)
	var nanny struct {
		ID            int
		Name          string
		Experience    string
		Phone         string
		AverageRating float64
		Description   string
		Price         float64
		PhotoURL      string
		City          string
	}

	err = row.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.AverageRating, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.City)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Няня не найдена", http.StatusNotFound)
			return
		} else {
			log.Println("Ошибка при выполнении запроса о няне:", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	// Получаем имя пользователя
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя", http.StatusInternalServerError)
		return
	}

	// Получаем роль пользователя
	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя", http.StatusInternalServerError)
		return
	}

	// Получаем отзывы о няне
	reviewQuery := "SELECT user_id, rating, comment, created_at FROM reviews WHERE nanny_id = $1"
	rows, err := Db.Query(reviewQuery, nanny.ID)
	if err != nil {
		log.Println("Ошибка при получении отзывов:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var reviews []struct {
		UserID    int
		Rating    float64
		Comment   string
		CreatedAt string
	}

	for rows.Next() {
		var review struct {
			UserID    int
			Rating    float64
			Comment   string
			CreatedAt string
		}
		if err := rows.Scan(&review.UserID, &review.Rating, &review.Comment, &review.CreatedAt); err != nil {
			log.Println("Error scanning review:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		reviews = append(reviews, review)
	}

	// Создаем структуру для передачи в шаблон
	pageData := struct {
		UserID   int
		UserName string
		Role     string
		Nanny    struct {
			ID            int
			Name          string
			Experience    string
			Phone         string
			AverageRating float64
			Description   string
			Price         float64
			PhotoURL      string
			City          string
		}
		Reviews []struct {
			UserID    int
			Rating    float64
			Comment   string
			CreatedAt string
		}
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Nanny:    nanny,
		Reviews:  reviews,
	}

	// Рендерим шаблон с данными
	tmpl, err := template.ParseFiles("templates/nanny_page.html")
	if err != nil {
		log.Println("Error loading template:", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		log.Printf("Error rendering template: %v, PageData: %+v", err, pageData)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// GetNannyByID возвращает информацию о няне по ее ID.
func GetNannyByID(nannyID int) (Nanny, error) {
	var nanny Nanny
	err := Db.QueryRow("SELECT id, name, experience, phone, description, price, photo_url, average_rating, review_count, city FROM nannies WHERE id = $1", nannyID).
		Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.AverageRating, &nanny.ReviewCount, &nanny.City)
	if err != nil {
		return nanny, err
	}
	return nanny, nil
}

// NannyPage отображает страницу с информацией о няне.
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
	err = Db.QueryRow("SELECT id, name, description, price, photo_url, city FROM nannies WHERE id = $1", nannyID).Scan(&nanny.ID, &nanny.Name, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.City)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Няня не найдена", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка при получении информации о няне", http.StatusInternalServerError)
		}
		return
	}

	// Получаем отзывы о няне
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

// UpdateNannyHandler обновляет профиль няни.
func UpdateNannyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Получение данных из формы
		nannyID := r.FormValue("id")
		name := r.FormValue("name")
		description := r.FormValue("description")
		priceStr := r.FormValue("price")
		city := r.FormValue("city")
		photoURL := r.FormValue("photo_url")

		// Преобразование цены из строки в float64
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			http.Error(w, "Неверный формат цены", http.StatusBadRequest)
			return
		}

		// Обновление данных о няне в базе данных
		query := "UPDATE nannies SET name = $1, description = $2, price = $3, photo_url = $4, city = $5 WHERE id = $6"
		_, err = Db.Exec(query, name, description, price, photoURL, city, nannyID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении профиля няни", http.StatusInternalServerError)
			return
		}

		// Перенаправление обратно на страницу с каталогом или панелью
		http.Redirect(w, r, "/edit_nanny", http.StatusFound)
	} else {
		http.Error(w, "Неверный метод запроса", http.StatusMethodNotAllowed)
	}
}

// GetNanniesWithRatings возвращает список нянь с их рейтингами.
func GetNanniesWithRatings() ([]Nanny, error) {
	var nannies []Nanny

	// Запрос для получения данных о нянях
	rows, err := Db.Query("SELECT id, name, experience, phone, description, price, photo_url, city FROM nannies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var nanny Nanny
		err := rows.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.City)
		if err != nil {
			return nil, err
		}

		// Получаем рейтинг и количество отзывов для этой няни
		ratings, err := GetRatingsForNanny(nanny.ID)
		if err != nil {
			return nil, err
		}

		// Рассчитываем средний рейтинг и количество отзывов
		nanny.ReviewCount = len(ratings)
		nanny.AverageRating = CalculateAverageRating(ratings)

		nannies = append(nannies, nanny)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nannies, nil
}

// GetReviewsByNannyID получает отзывы о няне по ее ID.
func GetReviewsByNannyID(nannyID int) ([]Review, error) {
	var reviews []Review

	// Запрос на получение отзывов
	query := "SELECT review_id, user_id, nanny_id, rating, comment, created_at FROM reviews WHERE nanny_id = $1"
	rows, err := Db.Query(query, nannyID)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Считываем отзывы
	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ReviewID, &review.UserID, &review.NannyID, &review.Rating, &review.Comment, &review.CreatedAt); err != nil {
			log.Printf("Ошибка при сканировании отзыва: %v", err)
			return nil, err
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации по строкам: %v", err)
		return nil, err
	}

	return reviews, nil
}

func EditNannyHandler(w http.ResponseWriter, r *http.Request) {
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

	// Выводим userID для отладки
	log.Print("UserID:", userID)

	// Предполагается, что ID няни передается через URL или форму,
	// например, вы можете получить его через параметр URL, например /edit_nanny?id=1

	// Получение данных о няне из базы данных
	nanny, err := GetNannyByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Няня не найдена", http.StatusNotFound)
			return
		}
		log.Println("Ошибка получения данных о няне:", err)
		http.Error(w, "Ошибка получения данных о няне", http.StatusInternalServerError)
		return
	}

	// Отправка данных в шаблон редактирования
	data := struct {
		UserID int
		Nanny  Nanny
	}{
		UserID: userID,
		Nanny:  nanny,
	}

	// Замените "templates/edit_nanny.html" на ваш путь к шаблону
	err = TmplEditNanny.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Функция для расчета среднего рейтинга
func CalculateAverageRating(ratings []float64) float64 {
	if len(ratings) == 0 {
		return 0 // Если рейтингов нет, возвращаем 0
	}

	var total float64
	for _, rating := range ratings {
		total += rating
	}
	return total / float64(len(ratings))
}
