package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"text/template"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Структура пользователя
type User struct {
	IDuser   int
	Login    string
	Password string
	Role     string
}

// Добавьте эту структуру
type NannyDetailPage struct {
	UserID   int
	UserName string
	Role     string
	Nanny    Nanny
}

type PageData struct {
	UserID   int
	UserName string
	Role     string
	Nannies  []Nanny // Добавляем список нянь
}

type Nanny struct {
	ID          int
	Name        string
	Role        string
	Experience  string // Добавлено поле для опыта
	Phone       string // Добавлено поле для телефона
	Description string
	Price       float64
	PhotoURL    string
}

var Db *sql.DB
var TmplRegister = template.Must(template.ParseFiles("templates/register.html"))
var TmplMain = template.Must(template.ParseFiles("templates/upload1.html"))
var TmplAdmin = template.Must(template.ParseFiles("templates/admin_page.html"))
var TmplNanny = template.Must(template.ParseFiles("templates/nanny_page.html"))
var TmplHome = template.Must(template.ParseFiles("templates/home.html"))

// Функция для открытия базы данных
func OpenDatabase() {
	var err error
	Db, err = sql.Open("postgres", "user=postgres password=1234 dbname=airat sslmode=disable")
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}
	err = Db.Ping()
	if err != nil {
		log.Fatal("Не удалось выполнить ping базы данных:", err)
	}
	log.Println("Успешно подключено к базе данных")
}

// Получаем имя пользователя по ID из базы данных
func GetUserNameByIDFromDB(userID int) (string, error) {
	var userName string
	query := "SELECT login FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя из базы данных:", err)
		return "", err
	}
	return userName, nil
}

// Получаем роль пользователя по ID из базы данных
func GetUserRoleByIDFromDB(userID int) (string, error) {
	var role string
	query := "SELECT role FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&role)
	if err != nil {
		log.Println("Ошибка при получении роли пользователя из базы данных:", err)
		return "", err
	}
	return role, nil
}

// Обработчик главной страницы
// Обработчик главной страницы
func Index(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
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

	// Добавляем данные для шаблона
	data := struct {
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	// Отправляем данные в шаблон
	err = TmplHome.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик страницы входа и регистрации
func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		TmplRegister.Execute(w, nil) // Отображение страницы регистрации/входа
	} else if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")

		// Попробуем аутентифицировать пользователя
		id := AuthenticateUser(r.Context(), login, password)

		if id == -1 {
			// Пользователь существует, но пароль неверный
			log.Println("Ошибка аутентификации: неверный пароль")
			http.ServeFile(w, r, "templates/errorModal.html") // Ошибка при аутентификации
		} else if id > 0 {
			// Пользователь успешно аутентифицирован
			log.Printf("Успешная аутентификация для пользователя с ID: %d", id)
			http.Redirect(w, r, "/main?id="+strconv.Itoa(id), http.StatusFound)
		} else {
			// Пользователь не найден — создаем новую учетную запись
			log.Println("Регистрация нового пользователя")
			hashedPassword, err := HashPassword(password)
			if err != nil {
				http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
				return
			}

			// Вставляем нового пользователя в базу данных
			query := "INSERT INTO users (login, password, role) VALUES ($1, $2, 'user') RETURNING id"
			var newUserID int
			err = Db.QueryRow(query, login, hashedPassword).Scan(&newUserID)
			if err != nil {
				http.Error(w, "Ошибка при регистрации пользователя", http.StatusInternalServerError)
				return
			}

			log.Printf("Пользователь успешно зарегистрирован с ID: %d", newUserID)
			// После успешной регистрации автоматически заходим в систему
			http.Redirect(w, r, "/main?id="+strconv.Itoa(newUserID), http.StatusFound)
		}
	}
}

// Обработчик для регистрации
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Отображение страницы регистрации
		TmplRegister.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		// Обработка формы регистрации
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Ошибка при обработке формы", http.StatusInternalServerError)
			return
		}
		login := r.Form.Get("login")
		password := r.Form.Get("password")

		hashedPassword, err := HashPassword(password)
		if err != nil {
			http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
			return
		}

		query := "INSERT INTO users (login, password, role) VALUES ($1, $2, 'user')" // Роль по умолчанию
		_, err = Db.Exec(query, login, hashedPassword)
		if err != nil {
			http.Error(w, "Ошибка при вставке пользователя", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

// Функция аутентификации
func AuthenticateUser(ctx context.Context, login, password string) int {
	query := "SELECT id, password, role FROM users WHERE login = $1"
	var currentUser User
	var hashedPassword string
	err := Db.QueryRow(query, login).Scan(&currentUser.IDuser, &hashedPassword, &currentUser.Role)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return 0 // Если пользователя не найдено, возвращаем 0
	}
	if CheckPasswordHash(password, hashedPassword) {
		return currentUser.IDuser // Возвращаем ID пользователя при успешной аутентификации
	} else {
		log.Println("Неверный пароль для пользователя:", login)
		return -1 // Возвращаем -1 при неверном пароле
	}
}

// Хеширование пароля
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Проверка пароля
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// / Получаем список всех пользователей
func GetAllUsers() ([]User, error) {
	query := "SELECT id, login, role FROM users"
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.IDuser, &user.Login, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

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

func GetNannies() []Nanny {
	rows, err := Db.Query("SELECT id, name, description, price, photo_url FROM nannies")
	if err != nil {
		log.Fatal("Ошибка при получении списка нянь из базы данных:", err)
	}
	defer rows.Close()

	var nannies []Nanny
	for rows.Next() {
		var nanny Nanny
		err := rows.Scan(&nanny.ID, &nanny.Name, &nanny.Description, &nanny.Price, &nanny.PhotoURL)
		if err != nil {
			log.Println("Ошибка при сканировании данных няни:", err)
			continue
		}
		nannies = append(nannies, nanny)
	}

	if err := rows.Err(); err != nil {
		log.Println("Ошибка при обработке строк с няньями:", err)
	}

	return nannies
}

// Структура для нянь

// Функция для парсинга HTML-шаблонов
func ParseTemplate(filename string) *template.Template {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Fatal(err)
	}
	return tmpl
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
	nannies := GetNannies()

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

// Обработчик для страницы Няни
func NannyPage(w http.ResponseWriter, r *http.Request) {
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

	// Получаем список нянь из базы данных
	nannies := GetNannies()

	// Подготовка данных для передачи в шаблон
	data := struct {
		UserID   int
		UserName string
		Role     string
		Nannies  []Nanny // Добавляем список нянь
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Nannies:  nannies,
	}

	// Выполняем рендеринг шаблона с обновленной структурой данных
	err = TmplNanny.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func NannyHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор няньки из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing nanny ID", http.StatusBadRequest)
		return
	}

	// Получение информации о няне из базы данных
	query := "SELECT id, name, experience, phone, description, price, photo_url FROM nannies WHERE id = $1"
	row := Db.QueryRow(query, idStr)

	var nanny Nanny
	err := row.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Nanny not found", http.StatusNotFound)
		} else {
			log.Println("Error querying nanny:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	// Здесь вы можете получить данные о текущем пользователе (например, UserID, UserName, Role)
	// Предположим, что вы передаете UserID через параметры запроса (добавьте это в URL для тестирования)
	userID, err := strconv.Atoi(r.URL.Query().Get("user_id")) // Получаем UserID из параметров запроса
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

	// Создаем структуру для передачи в шаблон
	pageData := NannyDetailPage{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Nanny:    nanny,
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
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
