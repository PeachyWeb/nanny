package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Обработчик для страницы пользователя
func UserPage(w http.ResponseWriter, r *http.Request) {
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

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя:", err)
		http.Error(w, "Ошибка при получении имени пользователя", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении роли пользователя:", err)
		http.Error(w, "Ошибка при получении роли пользователя", http.StatusInternalServerError)
		return
	}

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

	err = TmplMain.Execute(w, data)
	if err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}

// Карта для перевода дней недели на русский
var weekdays = map[time.Weekday]string{
	time.Monday:    "Понедельник",
	time.Tuesday:   "Вторник",
	time.Wednesday: "Среда",
	time.Thursday:  "Четверг",
	time.Friday:    "Пятница",
	time.Saturday:  "Суббота",
	time.Sunday:    "Воскресенье",
}

func CatalogPage(w http.ResponseWriter, r *http.Request) {
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

	// Получаем имя и роль пользователя из базы данных
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя из базы данных:", err)
		http.Error(w, "Ошибка при получении имени пользователя", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		log.Println("Ошибка при получении роли пользователя из базы данных:", err)
		http.Error(w, "Ошибка при получении роли пользователя", http.StatusInternalServerError)
		return
	}

	// Получаем параметры фильтрации из URL-запроса
	minExperience, maxPrice, minRating, err := parseFilters(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем список нянь с учетом фильтров
	nannies, err := GetNannies(Db, minExperience, maxPrice, minRating)
	if err != nil {
		log.Println("Ошибка при получении нянь из базы данных:", err)
		http.Error(w, "Ошибка при получении нянь из базы данных", http.StatusInternalServerError)
		return
	}

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
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}

// CalendarHandler обрабатывает запросы на страницу календаря
func CalendarHandler(w http.ResponseWriter, r *http.Request) {
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

	// Получаем месяц и год из параметров запроса
	monthIndexStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

	// Парсим выбранный месяц
	monthIndex, err := strconv.Atoi(monthIndexStr)
	if err != nil || monthIndex < 0 || monthIndex > 11 {
		monthIndex = int(time.Now().Month()) - 1 // Текущий месяц по умолчанию
	}

	// Парсим выбранный год
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		year = time.Now().Year() // Текущий год по умолчанию
	}

	// Получаем записи о встречах из базы данных
	query := `SELECT nannyid, starttime, endtime FROM appointments WHERE userid = $1`
	rows, err := Db.Query(query, userID)
	if err != nil {
		log.Println("Ошибка при запросе к appointments:", err)
		http.Error(w, "Unable to retrieve calendar data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.NannyID, &app.StartTime, &app.EndTime); err != nil {
			log.Println("Ошибка при сканировании записи о встрече:", err)
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		appointments = append(appointments, app)
	}

	username, err := GetUserNameByID(userID)
	if err != nil {
		http.Error(w, "Ошибка получения данных пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	busyDays := make(map[int]bool)
	for _, app := range appointments {
		if app.StartTime.Year() == year && int(app.StartTime.Month())-1 == monthIndex {
			day := app.StartTime.Day()
			busyDays[day] = true
		}
	}

	// Подготовка данных для шаблона
	data := struct {
		UserID             int
		UserName           string
		CurrentMonth       Month
		Months             []Month
		SelectedMonthIndex int
		Years              []int
		SelectedYear       int
	}{
		UserID:             userID,
		UserName:           username,
		SelectedMonthIndex: monthIndex,
		SelectedYear:       year,
		Months:             []Month{},
		Years:              generateYears(2020, 2030),
	}

	// Генерация месяцев для выбора
	for i := 0; i < 12; i++ {
		month := time.Month(i + 1)
		daysCount := daysInMonthInYear(month, year)
		daysInMonth := make([]Day, daysCount)
		for day := 1; day <= daysCount; day++ {
			date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			daysInMonth[day-1] = Day{
				Day:     day,
				Weekday: weekdays[date.Weekday()], // Устанавливаем название дня недели на русском
				IsBusy:  busyDays[day],
			}
		}
		data.Months = append(data.Months, Month{
			Name:  month.String(),
			Year:  year,
			Days:  daysInMonth,
			Index: i,
		})
	}

	// Добавляем данные для выбранного месяца
	daysCount := daysInMonthInYear(time.Month(monthIndex+1), year)
	daysInMonth := make([]Day, daysCount)
	for day := 1; day <= daysCount; day++ {
		date := time.Date(year, time.Month(monthIndex+1), day, 0, 0, 0, 0, time.UTC)
		daysInMonth[day-1] = Day{
			Day:     day,
			Weekday: weekdays[date.Weekday()],
			IsBusy:  busyDays[day],
		}
	}
	data.CurrentMonth = Month{
		Name:  time.Month(monthIndex + 1).String(),
		Year:  year,
		Days:  daysInMonth,
		Index: monthIndex,
	}

	// Рендерим шаблон
	tmpl := ParseTemplate("templates/calendar.html")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		http.Error(w, "Ошибка при генерации календаря", http.StatusInternalServerError)
	}
}

// Дополнительные функции и типы

type Month struct {
	Name  string
	Year  int
	Days  []Day
	Index int
}

// Day представляет структуру для дня с назначениями

// GetUserInfoByID возвращает информацию о клиенте по его ID
func GetUserInfoByID(userID int) (UserInfo, error) {
	var userInfo UserInfo
	query := `SELECT name, phone FROM users WHERE id = $1`
	err := Db.QueryRow(query, userID).Scan(&userInfo.Name, &userInfo.Phone)
	if err != nil {
		return UserInfo{}, err
	}
	return userInfo, nil
}

// UserInfo содержит данные о клиенте
type UserInfo struct {
	Name  string
	Phone string
}

// Функция для генерации списка месяцев
func generateMonths() []Month {
	months := []Month{}
	for i := 0; i < 12; i++ {
		month := time.Month(i + 1)
		months = append(months, Month{
			Name:  month.String(),
			Index: i,
		})
	}
	return months
}

// CheckIfUserIsNanny проверяет, является ли пользователь няней
func CheckIfUserIsNanny(userID int) (bool, error) {
	var isNanny bool
	// Проверяем наличие записи о няне в таблице nannies по user_id
	query := `SELECT 1 FROM nannies WHERE user_id = $1 LIMIT 1`
	err := Db.QueryRow(query, userID).Scan(&isNanny)
	if err != nil {
		// Если запись не найдена, то возвращаем ошибку
		if err == sql.ErrNoRows {
			// Пользователь не является няней
			return false, nil
		}
		// Ошибка при запросе к базе данных
		log.Println("Ошибка при проверке роли пользователя:", err)
		return false, err
	}
	// Если запись найдена, значит пользователь является няней
	return true, nil
}

// Функция для получения выбранного месяца и года из параметров запроса
func getSelectedMonthYear(r *http.Request) (int, int) {
	monthIndexStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

	// Парсим выбранный месяц
	monthIndex, err := strconv.Atoi(monthIndexStr)
	if err != nil || monthIndex < 0 || monthIndex > 11 {
		monthIndex = int(time.Now().Month()) - 1 // Текущий месяц по умолчанию
	}

	// Парсим выбранный год
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		year = time.Now().Year() // Текущий год по умолчанию
	}
	return monthIndex, year
}

// Функция для генерации списка годов
func generateYears(startYear, endYear int) []int {
	var years []int
	for year := startYear; year <= endYear; year++ {
		years = append(years, year)
	}
	return years
}

// daysInMonthInYear возвращает количество дней в заданном месяце и году
func daysInMonthInYear(month time.Month, year int) int {
	switch month {
	case time.January, time.March, time.May, time.July, time.August, time.October, time.December:
		return 31
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

// isLeapYear определяет, является ли год високосным
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func OrderDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

	dayStr := r.URL.Query().Get("day")
	monthStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

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

	// Для отладки выводим передаваемые параметры
	log.Printf("Полученные параметры: userID=%d, day=%d, month=%d, year=%d\n", userID, day, month, year)

	// SQL-запрос с выводом отладочного сообщения
	query := `SELECT a.idappointment, n.name, a.starttime, a.endtime 
	          FROM appointments a 
	          JOIN nannies n ON a.nannyid = n.id 
	          WHERE a.userid = $1 
	            AND EXTRACT(DAY FROM a.starttime) = $2 
	            AND EXTRACT(MONTH FROM a.starttime) = $3 
	            AND EXTRACT(YEAR FROM a.starttime) = $4`

	log.Printf("Выполняем SQL-запрос: %s с параметрами userID=%d, day=%d, month=%d, year=%d", query, userID, day, month+1, year)

	row := Db.QueryRow(query, userID, day, month+1, year)

	var orderID int
	var nannyName string
	var startTime, endTime time.Time
	err = row.Scan(&orderID, &nannyName, &startTime, &endTime)
	if err != nil {
		// Выводим сообщение об ошибке и SQL-параметры для отладки
		log.Printf("Ошибка при получении информации о заказе (возможно, данные не найдены): %v. Параметры запроса: userID=%d, day=%d, month=%d, year=%d", err, userID, day, month+1, year)
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	// Отладочный вывод для проверки значений времени
	log.Printf("Заказ ID: %d, Няня: %s, Начало: %s, Конец: %s\n", orderID, nannyName, startTime, endTime)
	// Подготовка данных для шаблона с полным форматом даты
	data := struct {
		OrderID   int
		NannyName string
		StartTime string
		EndTime   string
		Day       int
		Month     int
		Year      int
	}{
		OrderID:   orderID,
		NannyName: nannyName,
		StartTime: startTime.Format("02/01/2006, 15:04"),
		EndTime:   endTime.Format("02/01/2006, 15:04"),
		Day:       day,
		Month:     month,
		Year:      year,
	}

	// Выполнение шаблона для отображения заказа
	err = TmplOrderDetails.Execute(w, data)
	if err != nil {
		log.Println("Ошибка при выполнении шаблона:", err)
		http.Error(w, "Ошибка отображения страницы заказа", http.StatusInternalServerError)
	}
}

// getOrdersForNanny получает заказы для конкретной няни и подготавливает данные для шаблона
func getOrdersForNanny(nannyID int) (map[string][]Order, error) {
	rows, err := Db.Query(`
		SELECT a.idappointment, a.userid, a.nannyid, a.starttime, a.endtime, a.price,
		       u.first_name, u.last_name, u.patronymic, u.phone
		FROM appointments a
		JOIN users u ON a.userid = u.id
		WHERE a.nannyid = $1`, nannyID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	orders := map[string][]Order{
		"upcoming":  {},
		"completed": {},
	}

	for rows.Next() {
		var order Order
		var firstName, lastName, patronymic, phone sql.NullString
		var startTime, endTime time.Time

		if err := rows.Scan(&order.ID, &order.UserID, &order.NannyID, &startTime, &endTime, &order.Price,
			&firstName, &lastName, &patronymic, &phone); err != nil {
			return nil, err
		}

		order.FirstName = getValidString(firstName)
		order.LastName = getValidString(lastName)
		order.Patronymic = getValidString(patronymic)
		order.PhoneNumber = getValidString(phone)
		order.StartTime = formatTime(startTime)
		order.EndTime = formatTime(endTime)

		if endTime.After(now) {
			orders["upcoming"] = append(orders["upcoming"], order)
		} else {
			orders["completed"] = append(orders["completed"], order)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func formatTime(t time.Time) time.Time {
	return t.Truncate(time.Minute) // Округляем время до ближайшей минуты
}

func OrdersHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Ошибка при получении сессии", http.StatusInternalServerError)
		return
	}

	nannyID, ok := session.Values["userID"].(int)
	if !ok || nannyID <= 0 {
		http.Error(w, "Необходимо войти в систему", http.StatusUnauthorized)
		return
	}

	orders, err := getOrdersForNanny(nannyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем значения фильтров из запроса
	status := r.URL.Query().Get("status")
	name := r.URL.Query().Get("name")
	dateSort := r.URL.Query().Get("dateSort")

	// Фильтрация по статусу заказа
	if status == "upcoming" {
		orders["completed"] = nil
	} else if status == "completed" {
		orders["upcoming"] = nil
	}

	// Фильтрация по имени клиента
	if name != "" {
		for key := range orders {
			var filteredOrders []Order
			for _, order := range orders[key] {
				if strings.Contains(strings.ToLower(order.FirstName+" "+order.LastName), strings.ToLower(name)) {
					filteredOrders = append(filteredOrders, order)
				}
			}
			orders[key] = filteredOrders
		}
	}

	// Сортировка по дате начала
	for key := range orders {
		sort.Slice(orders[key], func(i, j int) bool {
			if dateSort == "oldest" {
				return orders[key][i].StartTime.Before(orders[key][j].StartTime)
			}
			return orders[key][i].StartTime.After(orders[key][j].StartTime)
		})
	}

	err = TmplOrders.Execute(w, struct {
		Orders map[string][]Order
	}{
		Orders: orders,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getValidString возвращает строку из sql.NullString, если она валидна
func getValidString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "Не указано" // Возвращаем текст, если значение не задано
}
