function Home({ userName, userID, nanny, reviews }) {
  return (
    <>
      <head>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Страница няни</title>
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

        {/* Информация о няне */}
        <div className="container mt-5">
          <h1>{nanny.name}</h1>
          <p>Опыт: {nanny.experience}</p>
          <p>Телефон: {nanny.phone}</p>
          <p>Средний рейтинг: {nanny.averageRating}</p>
          <p>Описание: {nanny.description}</p>
          <p>Цена: {nanny.price}</p>
          <img src={nanny.photoURL} alt={nanny.name} className="img-fluid" />

          {/* Форма для найма няни */}
          <h2 className="mt-4">Оформить найм няни</h2>
          <form action="/hire_nanny" method="POST" className="mb-4">
            <input type="hidden" name="user_id" value={userID} />
            <input type="hidden" name="nanny_id" value={nanny.id} />

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

          {/* Форма для добавления отзыва */}
          <h2>Оставить отзыв</h2>
          <form action="/add_review" method="POST">
            <input type="hidden" name="user_id" value={userID} />
            <input type="hidden" name="nanny_id" value={nanny.id} />

            <div className="mb-3">
              <label htmlFor="rating" className="form-label">
                Рейтинг (1-5):
              </label>
              <input
                type="number"
                id="rating"
                name="rating"
                min="1"
                max="5"
                className="form-control"
                required
              />
            </div>

            <div className="mb-3">
              <label htmlFor="comment" className="form-label">
                Комментарий:
              </label>
              <textarea
                id="comment"
                name="comment"
                className="form-control"
              ></textarea>
            </div>

            <button type="submit" className="btn btn-primary">
              Оставить отзыв
            </button>
          </form>

          {/* Список отзывов */}
          <h2>Отзывы</h2>
          {reviews && reviews.length > 0 ? (
            <ul>
              {reviews.map((review, index) => (
                <li key={index}>
                  <strong>Пользователь {review.userID}</strong> поставил
                  оценку: {review.rating}/5
                  <br />
                  <em>{review.comment}</em> <br />
                  <small>Дата: {review.createdAt}</small>
                </li>
              ))}
            </ul>
          ) : (
            <p>Пока нет отзывов</p>
          )}
        </div>

        {/* Подвал */}
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
