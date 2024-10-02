package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig *oauth2.Config
	vkOauthConfig     *oauth2.Config
	oauthStateString  = "randomstring" // Используется для предотвращения CSRF-атак
)

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

/*
	func init() {
		// Инициализация конфигурации Google OAuth
		googleOauthConfig = &oauth2.Config{
			RedirectURL:  "http://localhost:8080/callback",  // URL для обратного вызова Google
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),     // Ваш ID клиента Google
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"), // Ваш секрет клиента Google
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		}

		// Инициализация конфигурации VK OAuth
		vkOauthConfig = &oauth2.Config{
			ClientID:     os.Getenv("VK_CLIENT_ID"),           // Ваш ID клиента VK
			ClientSecret: os.Getenv("VK_CLIENT_SECRET"),       // Ваш секрет клиента VK
			RedirectURL:  "http://localhost:8080/callback/vk", // URL для обратного вызова VK
			Scopes:       []string{"email"},                   // Запрашиваемые разрешения
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://oauth.vk.com/authorize",
				TokenURL: "https://oauth.vk.com/access_token",
			},
		}
	}
*/

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
		TmplRegister.Execute(w, nil)
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

		query := "INSERT INTO users (login, password, role) VALUES ($1, $2, 'user')"
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
			http.Redirect(w, r, "/main?id="+strconv.Itoa(id), http.StatusFound)
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
			http.Redirect(w, r, "/main?id="+strconv.Itoa(newUserID), http.StatusFound)
		}
	}
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
	if r.FormValue("state") != oauthStateString {
		http.Error(w, "state invalid", http.StatusBadRequest)
		log.Println("Некорректное значение state")
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println("Не удалось получить токен:", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

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

	// Проверьте пользователя в базе данных и, если он не существует, создайте его
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

	http.Redirect(w, r, "/main?id="+strconv.Itoa(userID), http.StatusFound)
}

// Обработчик главной страницы
func Index(w http.ResponseWriter, r *http.Request) {
	log.Printf("Запрос к Index: %s %s", r.Method, r.URL.Path)
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

	data := struct {
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	err = TmplHome.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}
func VKLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := vkOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func VKCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauthStateString {
		http.Error(w, "state invalid", http.StatusBadRequest)
		return
	}

	token, err := vkOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println("Не удалось получить токен:", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Используем токен для запроса информации о пользователе
	client := vkOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.vk.com/method/users.get?fields=email&access_token=" + token.AccessToken + "&v=5.131")
	if err != nil {
		http.Error(w, "Не удалось получить информацию о пользователе", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Response []struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Не удалось декодировать ответ от VK", http.StatusInternalServerError)
		return
	}

	if len(userInfo.Response) == 0 {
		http.Error(w, "Не удалось получить информацию о пользователе", http.StatusInternalServerError)
		return
	}

	log.Printf("Пользователь аутентифицирован: %+v", userInfo.Response[0])

	// Проверяем пользователя в базе данных и создаем нового, если его нет
	var userID int
	query := "SELECT id FROM users WHERE login = $1"
	err = Db.QueryRow(query, userInfo.Response[0].Email).Scan(&userID)

	if err == sql.ErrNoRows {
		// Пользователь не существует, создаем нового
		query = "INSERT INTO users (login, password, role) VALUES ($1, '', 'user') RETURNING id"
		err = Db.QueryRow(query, userInfo.Response[0].Email).Scan(&userID)
		if err != nil {
			http.Error(w, "Ошибка при регистрации пользователя через VK", http.StatusInternalServerError)
			return
		}
		log.Printf("Пользователь успешно зарегистрирован через VK с ID: %d", userID)
	} else if err != nil {
		http.Error(w, "Ошибка при проверке пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/main?id="+strconv.Itoa(userID), http.StatusFound)
}
