package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  = "randomstring" // Используется для предотвращения CSRF-атак
	store             = sessions.NewCookieStore([]byte("super-secret-key"))
)

// Функция для обработки главной страницы и аутентификации
func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		TmplRegister.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")

		id := AuthenticateUser(r.Context(), login, password)
		if id == -1 {
			log.Println("Ошибка аутентификации: неверный пароль")
			http.ServeFile(w, r, "templates/errorModal.html")
		} else if id > 0 {
			log.Printf("Успешная аутентификация для пользователя с ID: %d", id)
			// Создаем и сохраняем сессию
			session, _ := store.Get(r, "session-name")
			session.Values["userID"] = id
			session.Save(r, w)

			http.Redirect(w, r, "/main", http.StatusFound)
		} else {
			log.Println("Регистрация нового пользователя")
			hashedPassword, err := HashPassword(password)
			if err != nil {
				http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
				return
			}

			query := "INSERT INTO users (login, password, role) VALUES ($1, $2, 'user') RETURNING id"
			var newUserID int
			err = Db.QueryRow(query, login, hashedPassword).Scan(&newUserID)
			if err != nil {
				http.Error(w, "Ошибка при регистрации пользователя", http.StatusInternalServerError)
				return
			}

			log.Printf("Пользователь успешно зарегистрирован с ID: %d", newUserID)
			// Создаем и сохраняем сессию
			session, _ := store.Get(r, "session-name")
			session.Values["userID"] = newUserID
			session.Save(r, w)

			http.Redirect(w, r, "/main", http.StatusFound)
		}
	}
}

func init() {
	// Проверка переменных окружения
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("Переменные окружения GOOGLE_CLIENT_ID и GOOGLE_CLIENT_SECRET должны быть установлены.")
	}

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback/google",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

}

// Функция для открытия базы данных
func OpenDatabase() {
	var err error
	Db, err = sql.Open("postgres", "user=postgres password=Tiqjt7o1992 dbname=airat sslmode=disable")
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
		// Здесь можно отобразить страницу регистрации
		// TmplRegister.Execute(w, nil) // Здесь нужно использовать ваш шаблон
	} else if r.Method == http.MethodPost {
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

		query := "INSERT INTO users (login, password, role) VALUES ($1, $2, 'user') RETURNING id"
		var newUserID int
		err = Db.QueryRow(query, login, hashedPassword).Scan(&newUserID)
		if err != nil {
			http.Error(w, "Ошибка при вставке пользователя", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/main?id="+strconv.Itoa(newUserID), http.StatusFound)
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
		return 0
	}
	if CheckPasswordHash(password, hashedPassword) {
		return currentUser.IDuser
	} else {
		log.Println("Неверный пароль для пользователя:", login)
		return -1
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

// Google OAuth Handlers
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Запрос к GoogleLoginHandler: %s %s", r.Method, r.URL.Path)
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	log.Println("Перенаправление на URL авторизации Google:", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Запрос к GoogleCallbackHandler: %s %s", r.Method, r.URL.Path)

	// Проверка состояния для защиты от CSRF-атак
	if r.FormValue("state") != oauthStateString {
		http.Error(w, "state invalid", http.StatusBadRequest)
		log.Println("Некорректное значение state")
		return
	}

	// Обмен кода на токен
	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println("Не удалось получить токен:", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Получение информации о пользователе
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "Не удалось получить информацию о пользователе", http.StatusInternalServerError)
		log.Println("Ошибка при получении информации о пользователе:", err)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Не удалось декодировать ответ от Google", http.StatusInternalServerError)
		log.Println("Ошибка при декодировании ответа от Google:", err)
		return
	}

	log.Printf("Пользователь аутентифицирован: %v", userInfo)

	// Проверяем пользователя в базе данных и создаем нового, если его нет
	var userID int
	query := "SELECT id FROM users WHERE login = $1"
	err = Db.QueryRow(query, userInfo.Email).Scan(&userID)

	if err == sql.ErrNoRows {
		// Пользователь не существует, создаем нового
		query = "INSERT INTO users (login, password, role) VALUES ($1, '', 'user') RETURNING id"
		err = Db.QueryRow(query, userInfo.Email).Scan(&userID)
		if err != nil {
			http.Error(w, "Ошибка при регистрации пользователя через Google", http.StatusInternalServerError)
			log.Println("Ошибка при регистрации пользователя через Google:", err)
			return
		}
		log.Printf("Пользователь успешно зарегистрирован через Google с ID: %d", userID)
	} else if err != nil {
		http.Error(w, "Ошибка при проверке пользователя", http.StatusInternalServerError)
		log.Println("Ошибка при проверке пользователя:", err)
		return
	}

	// Устанавливаем значение сессии для текущего пользователя
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Ошибка при получении сессии", http.StatusInternalServerError)
		return
	}
	session.Values["userID"] = userID // Устанавливаем ID пользователя в сессию
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Ошибка при сохранении сессии", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/main", http.StatusFound)
}

// Обработчик главной страницы
func Index(w http.ResponseWriter, r *http.Request) {
	log.Printf("Запрос к Index: %s %s", r.Method, r.URL.Path) // Логируем сам запрос

	// Получаем сессию
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Println("Ошибка при получении сессии:", err)
		http.Error(w, "Ошибка при получении сессии", http.StatusInternalServerError)
		return
	}

	// Проверяем, сохранен ли идентификатор пользователя в сессии
	userID, ok := session.Values["userID"].(int)
	if !ok || userID <= 0 {
		log.Println("Попытка доступа к главной странице без авторизации.")
		http.Error(w, "Необходимо войти в систему", http.StatusUnauthorized)
		return
	}

	// Получаем имя пользователя
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		log.Printf("Ошибка при получении имени пользователя из базы данных для ID %d: %v", userID, err)
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Получаем роль пользователя
	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		log.Printf("Ошибка при получении роли пользователя из базы данных для ID %d: %v", userID, err)
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	log.Printf("Пользователь с ID %d успешно получил доступ к главной странице. Имя: %s, Роль: %s", userID, userName, role)

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

	// Выполняем шаблон главной страницы с данными пользователя
	err = TmplHome.Execute(w, data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона для пользователя %d: %v", userID, err)
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Здесь могут быть ваши другие функции, такие как GetUserNameByIDFromDB и GetUserRoleByIDFromDB
