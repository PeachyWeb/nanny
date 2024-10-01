package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	IDuser   int
	Login    string
	Password string
	Role     string
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

// Обновление данных пользователя
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		userIDStr := r.FormValue("id")
		newRole := r.FormValue("new_role")

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
			return
		}

		query := "UPDATE users SET role = $1 WHERE id = $2"
		_, err = Db.Exec(query, newRole, userID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении роли пользователя", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin/employees", http.StatusFound)
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
	// Получаем идентификатор няни из URL
	idNanny := r.URL.Query().Get("nanny_id")
	if idNanny == "" {
		http.Error(w, "Missing nanny ID", http.StatusBadRequest)
		return
	}

	// Получаем информацию о няне из базы данных
	query := "SELECT id, name, experience, phone, description, price, photo_url, average_rating, review_count FROM nannies WHERE id = $1"
	row := Db.QueryRow(query, idNanny)

	var nanny Nanny
	err := row.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.AverageRating, &nanny.ReviewCount)
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

	// Получаем UserID из параметров запроса
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
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

	// Получение отзывов о няне
	reviews, err := GetReviewsByNannyID(nanny.ID)
	if err != nil {
		http.Error(w, "Error getting reviews from database", http.StatusInternalServerError)
		return
	}

	// Создаем структуру для передачи в шаблон
	pageData := NannyDetailPage{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Nanny:    nanny,
		Reviews:  reviews, // Добавляем отзывы в структуру
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
		http.Redirect(w, r, "/nanny_page?id="+nannyID, http.StatusFound)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Handler для редактирования профиля няни
func EditNannyHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Получение данных о няне из базы данных
	nanny, err := GetNannyByID(userID)
	if err != nil {
		http.Error(w, "Error getting nanny data from database", http.StatusInternalServerError)
		return
	}

	// Отправка данных в шаблон редактирования профиля
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
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

func AddReviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	userIDStr := r.FormValue("user_id")
	nannyIDStr := r.FormValue("nanny_id")
	ratingStr := r.FormValue("rating")
	comment := r.FormValue("comment")

	// Преобразование данных
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

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
	http.Redirect(w, r, fmt.Sprintf("/nanny/details?user_id=%d&nanny_id=%d", userID, nannyID), http.StatusSeeOther)
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

/*
// isDayBusy проверяет, занят ли указанный день

	func isDayBusy(day int, appointments []Appointment) bool {
		for _, app := range appointments {
			if app.StartTime.Day() == day {
				return true
			}
		}
		return false
	}
*/

/*func init() {
	// Parse templates
	TmplCalendar = template.Must(template.New("calendar.html").Funcs(template.FuncMap{
		"isDayBusy": isDayBusy,
	}).ParseFiles("calendar.html"))
}*/

// isDayBusy checks if a given day is busy based on the list of appointments.
/*func isDayBusy(day int, appointments []Appointment) bool {
	for _, app := range appointments {
		if app.StartTime.Day() == day {
			return true
		}
	}
	return false
}*/

// CalendarHandler обрабатывает запросы на страницу календаря

func CalendarHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch appointments from database
	query := `SELECT nannyid, starttime, endtime FROM appointments WHERE userid = $1`
	rows, err := Db.Query(query, userID)
	if err != nil {
		log.Println("Error querying appointments:", err)
		http.Error(w, "Unable to retrieve calendar data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.NannyID, &app.StartTime, &app.EndTime); err != nil {
			log.Println("Error scanning appointment:", err)
			http.Error(w, "Error processing data", http.StatusInternalServerError)
			return
		}
		appointments = append(appointments, app)
	}

	username, err := GetUserNameByID(userID)
	if err != nil {
		http.Error(w, "Error getting user data from database", http.StatusInternalServerError)
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
		UserID   string
		UserName string
		Days     []int
		BusyDays map[int]bool
	}{
		UserID:   userIDStr,
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
		log.Println("Error executing template:", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}

func HireNannyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		userID := r.FormValue("user_id")
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
		http.Redirect(w, r, "/calendar?id="+userID, http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
