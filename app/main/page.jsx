import React from 'react';

function Home({ userName, userID, userRole }) {
  return (
    <>
      <head>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Главная страница</title>
        <link
          rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.1.3/css/bootstrap.min.css"
        />
        <style>{`
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
        `}</style>
      </head>

      <body>
        {/* Навигационная панель */}
        <nav className="navbar navbar-expand-lg navbar-light bg-light">
          <div className="container">
            <a className="navbar-brand" href={`/main?id=${userID}`}>
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
                  <a className="nav-link" href={`/main?id=${userID}`}>
                    Главная
                  </a>
                </li>
                <li className="nav-item">
                  <a className="nav-link" href={`/catalog?id=${userID}`}>
                    Каталог нянь
                  </a>
                </li>
                <li className="nav-item">
                  <a className="nav-link" href={`/calendar?id=${userID}`}>
                    Календарь
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </nav>

        <div className="container mt-5">
          {/* Используйте состояние или пропсы для UserName и Role */}
          <h1>Welcome, {userName}!</h1>
          <h2>Your role: {userRole}</h2>

          {/* Здесь условная логика для роли пользователя */}
          {(() => {
            if (userRole === "Admin") {
              return (
                <>
                  <h2>Панель администратора</h2>
                  <form action="/admin_page" method="GET">
                    <input type="hidden" name="id" value={userID} />
                    <button type="submit" className="btn btn-secondary">
                      Перейти на админскую страницу
                    </button>
                  </form>
                </>
              );
            } else if (userRole === "nanny") {
              return (
                <>
                  <h2>Панель няни</h2>
                  <p>
                    Добро пожаловать, няня! Вы можете просмотреть каталоги нянь и
                    управлять своими услугами.
                  </p>
                  <form action="/edit_nanny" method="GET">
                    <input type="hidden" name="id" value={userID} />
                    <button type="submit" className="btn btn-primary">
                      Перейти в каталог нянь
                    </button>
                  </form>
                </>
              );
            } else {
              return <p>У вас нет доступа к этой панели.</p>;
            }
          })()}
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
