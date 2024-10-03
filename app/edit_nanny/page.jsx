function Home() {
    return (
      <>
        <head>
          <title>Страница изменения страницы няни</title>
          <link
            rel="stylesheet"
            href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css"
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
            <h2>Редактировать профиль няни</h2>
            <form action="/update_nanny" method="POST">
              <input type="hidden" name="id" value="{{ .UserID }}" />
              <div className="mb-3">
                <label htmlFor="name" className="form-label">
                  Имя:
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  className="form-control"
                  defaultValue="{{ .Nanny.Name }}"
                  required
                />
              </div>
              <div className="mb-3">
                <label htmlFor="description" className="form-label">
                  Описание:
                </label>
                <textarea
                  id="description"
                  name="description"
                  className="form-control"
                  defaultValue="{{ .Nanny.Description }}"
                  required
                ></textarea>
              </div>
              <div className="mb-3">
                <label htmlFor="price" className="form-label">
                  Цена:
                </label>
                <input
                  type="number"
                  id="price"
                  name="price"
                  className="form-control"
                  defaultValue="{{ .Nanny.Price }}"
                  required
                />
              </div>
              <div className="mb-3">
                <label htmlFor="photo_url" className="form-label">
                  URL фото:
                </label>
                <input
                  type="text"
                  id="photo_url"
                  name="photo_url"
                  className="form-control"
                  defaultValue="{{ .Nanny.PhotoURL }}"
                />
              </div>
              <button type="submit" className="btn btn-primary">
                Сохранить изменения
              </button>
            </form>
  
            <h2 className="mt-4">Оформить найм няни</h2>
            <form action="/hire_nanny" method="POST" className="mb-4">
              <input type="hidden" name="user_id" value="{{ .UserID }}" />
              <input type="hidden" name="nanny_id" value="{{ .UserID }}" />
  
              <div className="mb-3">
                <label htmlFor="start_time" className="form-label">
                  Время начала:
                </label>
                <input
                  type="datetime-local"
                  id="start_time"
                  name="start_time"
                  className="form-control"
                  required
                />
              </div>
  
              <div className="mb-3">
                <label htmlFor="end_time" className="form-label">
                  Время окончания:
                </label>
                <input
                  type="datetime-local"
                  id="end_time"
                  name="end_time"
                  className="form-control"
                  required
                />
              </div>
  
              <button type="submit" className="btn btn-success">
                Нанять няню
              </button>
            </form>
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
  