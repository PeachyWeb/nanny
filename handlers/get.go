package handlers

import (
	"database/sql"
	"log"
)

func GetNanniesWithRatings() ([]Nanny, error) {
	var nannies []Nanny

	// Запрос для получения данных о нянях
	rows, err := Db.Query("SELECT id, name, experience, phone, description, price, photo_url FROM nannies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var nanny Nanny
		err := rows.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL)
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

	return nannies, nil
}

// Handler для отображения профиля няни с отзывами

func GetReviewsByNannyID(nannyID int) ([]Review, error) {
	var reviews []Review

	// Убедитесь, что названия колонок совпадают с теми, что в вашей базе данных
	query := "SELECT review_id, nanny_id, user_id, rating, comment FROM reviews WHERE nanny_id = $1"
	rows, err := Db.Query(query, nannyID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ID, &review.NannyID, &review.UserID, &review.Rating, &review.Comment); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return reviews, nil
}

func GetNannyByID(nannyID int) (Nanny, error) {
	var nanny Nanny
	err := Db.QueryRow("SELECT id, name, experience, phone, description, price, photo_url, average_rating, review_count FROM nannies WHERE id = $1", nannyID).
		Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.AverageRating, &nanny.ReviewCount)
	return nanny, err
}

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
func GetNannies(db *sql.DB, minExperience int, maxPrice int, minRating int) ([]Nanny, error) {
	var nannies []Nanny

	// Базовый SQL-запрос
	query := `SELECT id, name, experience, phone, description, price, photo_url, average_rating, review_count FROM nannies WHERE 1=1`

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
		if err := rows.Scan(&nanny.ID, &nanny.Name, &nanny.Experience, &nanny.Phone, &nanny.Description, &nanny.Price, &nanny.PhotoURL, &nanny.AverageRating, &nanny.ReviewCount); err != nil {
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

// Функция для получения пользователя по ID
func GetUserByID(userID int) (User, error) {
	var user User
	err := Db.QueryRow("SELECT id, login, role FROM users WHERE id = $1", userID).Scan(&user.IDuser, &user.Login, &user.Role)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// Функция для получения пользователя по ID
func GetUserNameByID(userID int) (string, error) {
	var login string
	err := Db.QueryRow("SELECT login FROM users WHERE id = $1", userID).Scan(&login)
	if err != nil {
		return "", err
	}
	return login, nil
}
