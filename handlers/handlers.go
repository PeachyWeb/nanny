package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

// Структура пользователя

type Review struct {
	ID        int
	NannyID   int
	UserID    int
	Rating    int
	Comment   string
	CreatedAt time.Time // Добавлено поле
}

type Appointment struct {
	NannyID   int
	StartTime time.Time
	EndTime   time.Time
}

type User struct {
	IDuser     int
	Login      string
	Password   string
	Role       string
	FirstName  sql.NullString // Может быть NULL
	LastName   sql.NullString // Может быть NULL
	Patronymic sql.NullString // Может быть NULL
	City       sql.NullString // Может быть NULL
	Phone      sql.NullString // Может быть NULL
	Age        sql.NullInt64  // Может быть NULL
}

type NannyDetailPage struct {
	UserID        int
	UserName      string
	Role          string
	Nanny         Nanny
	Reviews       []Review
	AverageRating float64 // Убедитесь, что это поле присутствует
}

type PageData struct {
	UserID        int
	UserName      string
	Role          string
	Nannies       []Nanny // Добавляем список нянь
	AverageRating float64 // Новое поле
}

type Nanny struct {
	ID            int
	Name          string
	Experience    int
	Phone         string
	Description   string
	Price         float64
	PhotoURL      string
	AverageRating float64 // Новое поле
	ReviewCount   int     // Новое поле
	UserID        int
	UserName      string
}

var Db *sql.DB
var TmplRegister = template.Must(template.ParseFiles("templates/register.html"))
var TmplMain = template.Must(template.ParseFiles("templates/upload1.html"))
var TmplAdmin = template.Must(template.ParseFiles("templates/admin_page.html"))
var TmplNanny = template.Must(template.ParseFiles("templates/nanny_page.html"))
var TmplEditNanny = template.Must(template.ParseFiles("templates/edit_nanny.html"))
var TmplHome = template.Must(template.ParseFiles("templates/home.html"))
var TmplCalendar = template.Must(template.ParseFiles("templates/calendar.html"))
var TmplProfile = template.Must(template.ParseFiles("templates/profile.html"))

// Обновление данных пользователя
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
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

	if r.Method == http.MethodPost {
		// Получение данных из формы
		targetUserIDStr := r.FormValue("user_id")
		targetUserID, err := strconv.Atoi(targetUserIDStr)
		if err != nil {
			http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
			return
		}

		newLogin := r.FormValue("new_login")
		newPassword := r.FormValue("new_password")
		newRole := r.FormValue("new_role")

		// Проверяем, существует ли пользователь с указанным ID
		var existingUserID int
		err = Db.QueryRow("SELECT id FROM users WHERE id = $1", targetUserID).Scan(&existingUserID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Пользователь с указанным ID не найден", http.StatusNotFound)
			} else {
				log.Println("Ошибка при проверке пользователя:", err)
				http.Error(w, "Ошибка при проверке пользователя", http.StatusInternalServerError)
			}
			return
		}

		// Начинаем создание запроса на обновление
		updateFields := []string{}
		args := []interface{}{}
		argCounter := 1

		// Добавляем новые значения, если они указаны
		if newLogin != "" {
			updateFields = append(updateFields, "login = $"+strconv.Itoa(argCounter))
			args = append(args, newLogin)
			argCounter++
		}

		if newPassword != "" {
			hashedPassword, hashErr := HashPassword(newPassword)
			if hashErr != nil {
				log.Println("Ошибка при хешировании пароля:", hashErr)
				http.Error(w, "Ошибка при обновлении пароля", http.StatusInternalServerError)
				return
			}
			updateFields = append(updateFields, "password = $"+strconv.Itoa(argCounter))
			args = append(args, hashedPassword)
			argCounter++
		}

		if newRole != "" {
			updateFields = append(updateFields, "role = $"+strconv.Itoa(argCounter))
			args = append(args, newRole)
			argCounter++
		}

		// Проверяем, есть ли что обновлять
		if len(updateFields) == 0 {
			http.Error(w, "Нет данных для обновления", http.StatusBadRequest)
			return
		}

		// Собираем полный запрос
		query := "UPDATE users SET " + strings.Join(updateFields, ", ") + " WHERE id = $" + strconv.Itoa(argCounter)
		args = append(args, targetUserID)

		// Выполняем запрос
		_, err = Db.Exec(query, args...)
		if err != nil {
			log.Println("Ошибка при обновлении данных пользователя:", err)
			http.Error(w, "Ошибка при обновлении данных пользователя", http.StatusInternalServerError)
			return
		}

		// Перенаправляем после успешного обновления
		http.Redirect(w, r, "/admin/employees", http.StatusFound)
	} else {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// Функция для парсинга HTML-шаблонов
func ParseTemplate(filename string) *template.Template {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Fatal(err)
	}
	return tmpl
}

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
		http.Error(w, "Missing nanny ID", http.StatusBadRequest)
		return
	}

	// Получаем информацию о няне из базы данных
	query := "SELECT id, name, experience, phone, average_rating, description, price, photo_url FROM nannies WHERE id = $1"
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
	}

	err = row.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.AverageRating, &nanny.Description, &nanny.Price, &nanny.PhotoURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Nanny not found", http.StatusNotFound)
			return
		} else {
			log.Println("Error querying nanny:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Error getting user name", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Error getting user role", http.StatusInternalServerError)
		return
	}

	// Получаем отзывы о няне
	reviewQuery := "SELECT user_id, rating, comment, created_at FROM reviews WHERE nanny_id = $1"
	rows, err := Db.Query(reviewQuery, nanny.ID)
	if err != nil {
		log.Println("Error querying reviews:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
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

// Handler для обновления профиля няни
func UpdateNannyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Получение данных из формы
		nannyID := r.FormValue("id")
		name := r.FormValue("name")
		description := r.FormValue("description")
		priceStr := r.FormValue("price")
		photoURL := r.FormValue("photo_url")

		// Преобразование цены из строки в float64
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			http.Error(w, "Invalid price", http.StatusBadRequest)
			return
		}

		// Обновление данных о няне в базе данных
		query := "UPDATE nannies SET name = $1, description = $2, price = $3, photo_url = $4 WHERE id = $5"
		_, err = Db.Exec(query, name, description, price, photoURL, nannyID)
		if err != nil {
			http.Error(w, "Error updating nanny profile", http.StatusInternalServerError)
			return
		}

		// Перенаправление обратно на страницу с каталогом или панелью
		http.Redirect(w, r, "/edit_nanny", http.StatusFound)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
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

func GetNannyByID(nannyID int) (Nanny, error) {
	var nanny Nanny
	err := Db.QueryRow("SELECT id, name, experience, phone, description, price, photo_url, average_rating, review_count FROM nannies WHERE id = $1", nannyID).
		Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.AverageRating, &nanny.ReviewCount)
	if err != nil {
		return nanny, err // Возвращаем пустую структуру и ошибку
	}
	return nanny, nil // Возвращаем заполненную структуру и nil
}

func AddReviewHandler(w http.ResponseWriter, r *http.Request) {
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
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из форм
	nannyIDStr := r.FormValue("nanny_id")
	ratingStr := r.FormValue("rating")
	comment := r.FormValue("comment")

	nannyID, err := strconv.Atoi(nannyIDStr)
	if err != nil {
		http.Error(w, "Invalid nanny ID", http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		http.Error(w, "Invalid rating value", http.StatusBadRequest)
		return
	}

	// Добавляем отзыв в базу данных
	_, err = Db.Exec("INSERT INTO reviews (user_id, nanny_id, rating, comment, created_at) VALUES ($1, $2, $3, $4, NOW())",
		userID, nannyID, rating, comment)
	if err != nil {
		http.Error(w, "Error adding review to the database", http.StatusInternalServerError)
		return
	}

	// Обновление рейтинга няни
	err = UpdateNannyRating(nannyID)
	if err != nil {
		http.Error(w, "Error updating nanny rating", http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу профиля няни
	http.Redirect(w, r, fmt.Sprintf("/nanny/details?nanny_id=%d", nannyID), http.StatusSeeOther)
}

func UpdateNannyRating(nannyID int) error {
	// Получаем все отзывы для няни
	rows, err := Db.Query("SELECT rating FROM reviews WHERE nanny_id = $1", nannyID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var totalRating int
	var reviewCount int

	for rows.Next() {
		var rating int
		if err := rows.Scan(&rating); err != nil {
			return err
		}
		totalRating += rating
		reviewCount++
	}

	var averageRating float64
	if reviewCount > 0 {
		averageRating = float64(totalRating) / float64(reviewCount)
	}

	// Обновляем данные о няне
	_, err = Db.Exec("UPDATE nannies SET average_rating = $1, review_count = $2 WHERE id = $3",
		averageRating, reviewCount, nannyID)
	return err
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

func HireNannyHandler(w http.ResponseWriter, r *http.Request) {
	//Получаем сессию
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
	if r.Method == http.MethodPost {

		nannyID := r.FormValue("nanny_id")
		startTime := r.FormValue("start_time")
		endTime := r.FormValue("end_time")

		// Преобразование времени из строки в формат time.Time
		start, err := time.Parse("2006-01-02T15:04", startTime)
		if err != nil {
			http.Error(w, "Invalid start time format", http.StatusBadRequest)
			return
		}

		end, err := time.Parse("2006-01-02T15:04", endTime)
		if err != nil {
			http.Error(w, "Invalid end time format", http.StatusBadRequest)
			return
		}

		// Проверка на корректность временного диапазона
		if end.Before(start) {
			http.Error(w, "End time must be after start time", http.StatusBadRequest)
			return
		}

		// Вставка данных в таблицу appointments
		query := `INSERT INTO appointments (userid, nannyid, starttime, endtime) VALUES ($1, $2, $3, $4)`
		_, err = Db.Exec(query, userID, nannyID, start, end)
		if err != nil {
			log.Println("Error inserting appointment into database:", err)
			http.Error(w, "Failed to hire nanny. Please try again later.", http.StatusInternalServerError)
			return
		}

		// Перенаправление пользователя на страницу календаря после успешного добавления записи
		http.Redirect(w, r, "/calendar", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
