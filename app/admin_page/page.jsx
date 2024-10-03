function Home() {
  return (
    <>
      <head>
        <title>Страница администратора</title>
        <link
          rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.1.3/css/bootstrap.min.css"
        />
        <style>
          {`
            body {
              background-color: #f8f9fa;
            }

            .navbar-brand {
              font-weight: bold;
            }

            .card-img-top {
              height: 250px;
              object-fit: cover;
            }

            .card {
              transition: transform 0.2s;
            }

            .card:hover {
              transform: scale(1.05);
            }

            .filter-section {
              background-color: white;
              padding: 20px;
              border-radius: 5px;
              box-shadow: 0 0 15px rgba(0, 0, 0, 0.1);
              margin-bottom: 20px;
            }

            footer {
              position: fixed;
              bottom: 0;
              width: 100%;
              text-align: center;
              background-color: white;
              padding: 10px;
            }
          `}
        </style>
      </head>
      <body>
        {/* Навигационная панель */}
        <nav className="navbar navbar-expand-lg navbar-light bg-light">
          <div className="container">
            <a className="navbar-brand" href="/main?id={{ .UserID }}">
              Nanny Market
            </a>
            <button
              className="navbar-toggler"
              type="button"
              data-bs-toggle="collapse"
              data-bs-target="#navbarNav"
              aria-controls="navbarNav"
              aria-expanded="false"
              aria-label="Toggle navigation"
            >
              <span className="navbar-toggler-icon"></span>
            </button>
            <div className="collapse navbar-collapse" id="navbarNav">
              <ul className="navbar-nav ms-auto">
                <li className="nav-item">
                  <a className="nav-link" href="/main?id={{ .UserID }}">
                    Главная
                  </a>
                </li>
                <li className="nav-item">
                  <a className="nav-link" href="/catalog?id={{ .UserID }}">
                    Каталог нянь
                  </a>
                </li>
                <li className="nav-item">
                  <a className="nav-link" href="/calendar?id={{ .UserID }}">
                    Календарь
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </nav>

        <div className="container mt-4">
          <h1>Добро пожаловать на админскую страницу!</h1>
          <p>Здесь могут находиться функции и инструменты для администраторов.</p>

          <h2>Изменить данные пользователя</h2>
          <form action="/update_user" method="POST">
            <div className="mb-3">
              <label htmlFor="user_id" className="form-label">
                ID пользователя:
              </label>
              <input type="text" id="user_id" name="user_id" className="form-control" required />
            </div>

            <div className="mb-3">
              <label htmlFor="new_login" className="form-label">
                Новый логин:
              </label>
              <input type="text" id="new_login" name="new_login" className="form-control" />
            </div>

            <div className="mb-3">
              <label htmlFor="new_password" className="form-label">
                Новый пароль:
              </label>
              <input type="password" id="new_password" name="new_password" className="form-control" />
            </div>

            <div className="mb-3">
              <label htmlFor="new_role" className="form-label">
                Новая роль:
              </label>
              <select id="new_role" name="new_role" className="form-select">
                <option value="user">User</option>
                <option value="admin">Admin</option>
                <option value="nanny">Nanny</option>
              </select>
            </div>

            <button type="submit" className="btn btn-primary">Обновить данные</button>
          </form>

          <h2 className="mt-4">Список всех сотрудников</h2>
          <form action="/admin/employees" method="GET" className="mb-4">
            <input type="hidden" name="id" value="{{ .UserID }}" />
            <button type="submit" className="btn btn-secondary">Показать всех сотрудников</button>
          </form>

          {/* Здесь предполагается, что список сотрудников передается через пропсы или состояние */}
          {false ? ( // Замените на условие, проверяющее наличие пользователей, например, users.length > 0
            <>
              <h3>Сотрудники:</h3>
              <ul>
                {/* users.map((user) => (
                  <li key={user.IDuser}>ID: {user.IDuser}, Логин: {user.Login}, Роль: {user.Role}</li>
                )) */}
              </ul>
            </>
          ) : (
            <p>Список сотрудников еще не загружен.</p>
          )}
        </div>

        <footer className="bg-light py-3 text-center mt-auto">
          <div className="container">
            <p className="mb-0">&copy; 2024 Nanny Market. Все права защищены.</p>
          </div>
        </footer>
      </body>
    </>
  );
}

export default Home;
