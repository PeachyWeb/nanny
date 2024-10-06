package main

import (
	"log"
	"net/http"
	"uml/handlers"
)

func main() {
	handlers.OpenDatabase()   // Открываем подключение к базе данных
	defer handlers.Db.Close() // Закрываем соединение с базой данных при завершении работы

	// Проверьте, что все маршруты зарегистрированы только один раз
	http.HandleFunc("/", handlers.Home) // Обработчик для главной страницы
	http.HandleFunc("/main", handlers.Index)
	http.HandleFunc("/nanny", handlers.NannyPage)
	http.HandleFunc("/calendar", handlers.CalendarHandler)
	http.HandleFunc("/profile", handlers.ProfilePage)
	http.HandleFunc("/order-history", handlers.OrderHistoryHandler)

	http.HandleFunc("/order/{order_id}/review", handlers.AddReviewForOrderHandler)

	// Страницы для ролей
	http.HandleFunc("/admin_page", handlers.AdminPage)
	http.HandleFunc("/edit_nanny", handlers.EditNannyHandler) // Обработчик для редактирования профиля

	http.HandleFunc("/user_page", handlers.UserPage) // Добавляем маршрут для "User"
	http.HandleFunc("/nanny/details", handlers.NannyHandler)
	http.HandleFunc("/hire_nanny", handlers.HireNannyHandler)

	http.HandleFunc("/profile/update", handlers.UpdateProfile)
	http.HandleFunc("/update_user", handlers.UpdateUserHandler)
	http.HandleFunc("/update_nanny", handlers.UpdateNannyHandler) // Обработчик для обновления профиля

	// Маршруты
	http.HandleFunc("/register", handlers.RegisterHandler) // Обработчик для регистрации
	http.HandleFunc("/catalog", handlers.CatalogPage)      // Обработчик для каталога нянь

	// Дополнительные маршруты для функционала
	http.HandleFunc("/admin/employees", handlers.AdminEmployeesPage)

	http.HandleFunc("/add_review", handlers.AddReviewHandler) // Обработчик для добавления отзыва

	http.HandleFunc("/register_nanny", handlers.RegisterNanny)
	http.HandleFunc("/become_nanny_guide", handlers.GuideNanny)

	http.HandleFunc("/login/google", handlers.GoogleLoginHandler)
	http.HandleFunc("/callback/google", handlers.GoogleCallbackHandler)

	log.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
