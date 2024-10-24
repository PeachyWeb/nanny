package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

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
		Years:              generateYears(2020, 2030), // Генерация диапазона годов
	}

	// Генерация месяцев для выбора
	for i := 0; i < 12; i++ {
		month := time.Month(i + 1)
		daysCount := daysInMonthInYear(month, year)
		daysInMonth := make([]Day, daysCount)
		for day := 1; day <= daysCount; day++ {
			daysInMonth[day-1] = Day{Day: day, IsBusy: busyDays[day]}
		}
		data.Months = append(data.Months, Month{
			Name:  month.String(),
			Year:  year,
			Days:  daysInMonth,
			Index: i, // Записываем индекс месяца
		})
	}

	// Добавляем данные для выбранного месяца
	daysCount := daysInMonthInYear(time.Month(monthIndex+1), year)
	daysInMonth := make([]Day, daysCount)
	for day := 1; day <= daysCount; day++ {
		daysInMonth[day-1] = Day{Day: day, IsBusy: busyDays[day]}
	}
	data.CurrentMonth = Month{
		Name:  time.Month(monthIndex + 1).String(),
		Year:  year,
		Days:  daysInMonth,
		Index: monthIndex, // Записываем индекс выбранного месяца
	}

	// Выполнение шаблона
	err = TmplCalendar.Execute(w, data)
	if err != nil {
		log.Println("Ошибка при выполнении шаблона:", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
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

type Day struct {
	Day    int
	IsBusy bool
}

type Month struct {
	Name  string
	Year  int
	Days  []Day
	Index int // Добавляем поле для индекса месяца
}

// OrderDetailsHandler обрабатывает запросы для отображения информации о заказе
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

	// Получаем информацию о заказе из базы данных
	query := `SELECT n.name, a.starttime, a.endtime 
              FROM appointments a 
              JOIN nannies n ON a.nannyid = n.id 
              WHERE a.userid = $1 AND EXTRACT(DAY FROM a.starttime) = $2 
              AND EXTRACT(MONTH FROM a.starttime) = $3 
              AND EXTRACT(YEAR FROM a.starttime) = $4`
	row := Db.QueryRow(query, userID, day, month+1, year)

	var nannyName string
	var startTime, endTime time.Time
	err = row.Scan(&nannyName, &startTime, &endTime)
	if err != nil {
		log.Println("Ошибка при получении информации о заказе:", err)
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	// Подготовка данных для шаблона
	data := struct {
		NannyName string
		StartTime string
		EndTime   string
	}{
		NannyName: nannyName,
		StartTime: startTime.Format("15:04"),
		EndTime:   endTime.Format("15:04"),
	}

	// Выполнение шаблона для отображения заказа
	err = TmplOrderDetails.Execute(w, data)
	if err != nil {
		log.Println("Ошибка при выполнении шаблона:", err)
		http.Error(w, "Ошибка отображения страницы заказа", http.StatusInternalServerError)
	}
}
