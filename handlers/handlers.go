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
	ReviewID  int
	UserID    int
	NannyID   int
	Rating    int
	Comment   string
	CreatedAt time.Time
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
	City          string
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
var TmplNannyGuide = template.Must(template.ParseFiles("templates/nanny_guid.html"))

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

// RegisterNanny обрабатывает регистрацию будущей няни
func GuideNanny(w http.ResponseWriter, r *http.Request) {
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
		// Получаем данные из формы
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		patronymic := r.FormValue("patronymic")
		city := r.FormValue("city")
		phone := r.FormValue("phone")
		age := r.FormValue("age")

		// Сохраняем данные в базе данных
		_, err := Db.Exec("INSERT INTO future_nannies (user_id, first_name, last_name, patronymic, city, phone, age) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			userID, firstName, lastName, patronymic, city, phone, age)
		if err != nil {
			log.Println("Ошибка при регистрации няни:", err)
			http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу профиля
		http.Redirect(w, r, "/main", http.StatusSeeOther)
		return
	}

	// Обрабатываем GET-запрос, чтобы отобразить форму
	w.Header().Set("Content-Type", "text/html")
	if err := TmplNannyGuide.Execute(w, nil); err != nil { // Предполагается, что у вас есть шаблон TmplNannyGuide
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}

// RegisterNanny обрабатывает регистрацию будущей няни
func RegisterNanny(w http.ResponseWriter, r *http.Request) {
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
		// Получаем данные из формы
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		patronymic := r.FormValue("patronymic")
		city := r.FormValue("city")
		phone := r.FormValue("phone")
		age := r.FormValue("age")

		// Проверяем наличие справки о здоровье и релевантного опыта
		healthCertificate := r.FormValue("health_certificate") == "on"   // Это проверка на "on" или "off"
		relevantExperience := r.FormValue("relevant_experience") == "on" // То же самое

		// Сохраняем данные в базе данных
		_, err := Db.Exec(
			"INSERT INTO future_nannies (user_id, first_name, last_name, patronymic, city, phone, age, health_certificate, relevant_experience) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
			userID, firstName, lastName, patronymic, city, phone, age, healthCertificate, relevantExperience,
		)
		if err != nil {
			log.Println("Ошибка при регистрации няни:", err)
			http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу профиля
		http.Redirect(w, r, "/main", http.StatusSeeOther)
		return
	}

	http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
}
