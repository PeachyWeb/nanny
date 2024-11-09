package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func OrderHistoryHandler(w http.ResponseWriter, r *http.Request) {
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

	// Получение параметров из URL для фильтрации
	nannyName := r.URL.Query().Get("nannyName")
	dateFrom := r.URL.Query().Get("dateFrom")
	dateTo := r.URL.Query().Get("dateTo")
	sortBy := r.URL.Query().Get("sortBy") // Параметр для сортировки

	// Построение SQL-запроса с фильтрами
	query := `
        SELECT 
            a.idappointment, a.starttime, a.price, n.id, n.name,
            EXISTS(SELECT 1 FROM reviews WHERE nanny_id = n.id AND user_id = $1) AS review_left,
            r.comment, r.rating, r.created_at
        FROM appointments a
        JOIN nannies n ON a.nannyid = n.id
        LEFT JOIN reviews r ON r.nanny_id = n.id AND r.user_id = $1
        WHERE a.userid = $1
    `
	args := []interface{}{userID}
	argIndex := 2 // Начинаем с индекса 2, так как userID уже занят $1

	// Добавление условий в зависимости от параметров фильтрации
	if nannyName != "" {
		query += " AND n.name ILIKE $" + fmt.Sprint(argIndex)
		args = append(args, "%"+nannyName+"%")
		argIndex++
	}
	if dateFrom != "" {
		query += " AND a.starttime >= $" + fmt.Sprint(argIndex)
		args = append(args, dateFrom)
		argIndex++
	}
	if dateTo != "" {
		query += " AND a.starttime <= $" + fmt.Sprint(argIndex)
		args = append(args, dateTo)
		argIndex++
	}

	// Добавляем сортировку по запросу пользователя
	switch sortBy {
	case "dateAsc":
		query += " ORDER BY a.starttime ASC"
	case "dateDesc":
		query += " ORDER BY a.starttime DESC"
	case "priceAsc":
		query += " ORDER BY a.price ASC"
	case "priceDesc":
		query += " ORDER BY a.price DESC"
	default:
		query += " ORDER BY a.starttime DESC" // Сортировка по умолчанию
	}

	// Выполнение SQL-запроса
	rows, err := Db.Query(query, args...)
	if err != nil {
		log.Println("Ошибка при выполнении запроса к базе данных:", err)
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Обработка результатов запроса
	var orders []Order
	for rows.Next() {
		var order Order
		var reviewLeft bool
		var comment sql.NullString
		var rating sql.NullInt64
		var createdAt sql.NullTime

		err := rows.Scan(&order.ID, &order.StartTime, &order.Price, &order.NannyID, &order.NannyName,
			&reviewLeft, &comment, &rating, &createdAt)
		if err != nil {
			log.Println("Ошибка при сканировании заказов:", err)
			http.Error(w, "Ошибка при обработке заказов", http.StatusInternalServerError)
			return
		}

		order.ReviewLeft = reviewLeft

		// Если отзыв присутствует, создаем его
		if comment.Valid {
			order.Review = &Review{
				Comment:   comment.String,
				Rating:    int(rating.Int64), // Преобразование в int
				CreatedAt: createdAt.Time,
			}
		} else {
			order.Review = nil
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка при обработке данных:", err)
		http.Error(w, "Ошибка при обработке данных", http.StatusInternalServerError)
		return
	}

	// Формируем данные для шаблона
	pageData := struct {
		UserID int
		Orders []Order
	}{
		UserID: userID,
		Orders: orders,
	}

	// Рендеринг шаблона с переданными данными
	err = TmplOrderHistory.Execute(w, pageData)
	if err != nil {
		log.Println("Ошибка при рендеринге шаблона:", err)
		http.Error(w, "Ошибка при рендеринге шаблона", http.StatusInternalServerError)
	}
}

func CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Println("Ошибка при получении сессии:", err)
		http.Error(w, "Ошибка при получении сессии", http.StatusInternalServerError)
		return
	}

	userID, ok := session.Values["userID"].(int)
	if !ok || userID <= 0 {
		http.Error(w, "Необходимо войти в систему", http.StatusUnauthorized)
		return
	}

	// Получение параметров даты и orderID через FormValue для POST-запроса
	dayStr := r.FormValue("day")
	monthStr := r.FormValue("month")
	yearStr := r.FormValue("year")
	orderIDStr := r.FormValue("orderID")

	// Преобразование строковых параметров в целочисленные
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		http.Error(w, "Неверный день", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		http.Error(w, "Неверный месяц", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Неверный год", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Неверный ID заказа", http.StatusBadRequest)
		return
	}

	// Логирование полученных параметров для отладки
	log.Printf("Полученные параметры: userID=%d, day=%d, month=%d, year=%d, orderID=%d\n", userID, day, month, year, orderID)

	// Удаление заказа из базы данных
	query := `DELETE FROM appointments WHERE idappointment = $1 AND userid = $2`
	_, err = Db.Exec(query, orderID, userID)
	if err != nil {
		log.Println("Ошибка при отмене заказа:", err)
		http.Error(w, "Ошибка при отмене заказа", http.StatusInternalServerError)
		return
	}

	// Перенаправление на календарь после успешного удаления
	http.Redirect(w, r, "/calendar", http.StatusSeeOther)
}

// Обработчик для добавления отзыва
func AddReviewForOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

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

	// Получаем идентификатор заказа из параметров URL
	vars := mux.Vars(r)
	orderIDStr, ok := vars["orderID"]
	if !ok {
		http.Error(w, "ID заказа не найден", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID заказа", http.StatusBadRequest)
		return
	}

	// Проверяем, что отзыв для данного заказа ещё не оставлен
	var existingReviewCount int
	err = Db.QueryRow("SELECT COUNT(*) FROM reviews WHERE user_id = $1 AND order_id = $2", userID, orderID).Scan(&existingReviewCount)
	if err != nil {
		log.Println("Ошибка при проверке отзыва:", err)
		http.Error(w, "Ошибка при проверке отзыва", http.StatusInternalServerError)
		return
	}

	if existingReviewCount > 0 {
		http.Error(w, "Отзыв для этого заказа уже оставлен", http.StatusBadRequest)
		return
	}

	// Получаем данные из формы
	ratingStr := r.FormValue("rating")
	comment := r.FormValue("comment")

	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		http.Error(w, "Некорректное значение рейтинга", http.StatusBadRequest)
		return
	}

	// Добавляем отзыв в базу данных
	_, err = Db.Exec(`
        INSERT INTO reviews (user_id, nanny_id, rating, comment, created_at)
        VALUES ($1, (SELECT nannyid FROM appointments WHERE idappointment = $2), $3, $4, NOW())`,
		userID, orderID, rating, comment)
	if err != nil {
		log.Println("Ошибка при добавлении отзыва:", err)
		http.Error(w, "Ошибка при добавлении отзыва", http.StatusInternalServerError)
		return
	}

	// Перенаправление после успешного добавления отзыва
	http.Redirect(w, r, "/order-history", http.StatusSeeOther)
}
