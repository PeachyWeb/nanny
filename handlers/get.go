package handlers

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

func GetUserRoleByIDFromDB(userID int) (string, error) {
	var role string
	query := "SELECT role FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func GetUserNameByIDFromDB(userID int) (string, error) {
	var userName string
	query := "SELECT login FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		return "", err
	}
	return userName, nil
}
