package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

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
