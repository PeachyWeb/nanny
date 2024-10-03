import React from 'react';

function Home({ userID, nannies = [] }) { // Defaulting nannies to an empty array
  return (
    <>
      <head>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Каталог нянь</title>
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

        {/* Основной контент */}
        <div className="container py-5">
          {/* Фильтры */}
          <div className="row filter-section">
            <div className="col-md-3">
              <label htmlFor="filterExperience" className="form-label">
                Стаж (лет)
              </label>
              <input
                type="number"
                id="filterExperience"
                className="form-control"
                placeholder="Мин. стаж"
              />
            </div>
            <div className="col-md-3">
              <label htmlFor="filterPrice" className="form-label">
                Цена (руб./час)
              </label>
              <input
                type="number"
                id="filterPrice"
                className="form-control"
                placeholder="Макс. цена"
              />
            </div>
            <div className="col-md-3">
              <label htmlFor="filterRating" className="form-label">
                Рейтинг
              </label>
              <input
                type="number"
                id="filterRating"
                className="form-control"
                step="0.1"
                placeholder="Мин. рейтинг"
              />
            </div>
            <div className="col-md-3 d-flex align-items-end">
              <button id="applyFilters" className="btn btn-primary w-100">
                Применить фильтры
              </button>
            </div>
          </div>

          {/* Карточки нянь */}
          <div className="row g-4" id="nannyCatalog">
            {nannies.length > 0 ? ( // Check if nannies array is not empty
              nannies.map((nanny) => (
                <div
                  key={nanny.id}
                  className="col-12 col-md-6 col-lg-4 nanny-card"
                  data-experience={nanny.experience}
                  data-price={nanny.price}
                  data-rating={nanny.averageRating}
                >
                  <div className="card h-100 shadow-sm">
                    <a
                      href={`/nanny/details?user_id=${userID}&nanny_id=${nanny.id}`}
                      className="stretched-link text-decoration-none text-dark"
                    >
                      <img
                        src={nanny.photoURL}
                        className="card-img-top"
                        alt="Фото няни"
                      />
                      <div className="card-body">
                        <h5 className="card-title">{nanny.name}</h5>
                        <p className="card-text">{nanny.description}</p>
                        <p className="text-muted">{nanny.price} руб./час</p>
                        <p>Стаж: {nanny.experience}</p>
                        <p>Рейтинг: {nanny.averageRating}</p>
                      </div>
                    </a>
                  </div>
                </div>
              ))
            ) : (
              <div className="col-12">
                <p className="text-center">Няни не найдены.</p>
              </div>
            )}
          </div>
        </div>

        {/* Подвал */}
        <footer className="bg-light py-3 text-center mt-auto">
          <div className="container">
            <p className="mb-0">&copy; 2024 Nanny Market. Все права защищены.</p>
          </div>
        </footer>

        {/* JavaScript для фильтрации */}
        <script type="text/javascript">{`
          document.addEventListener('DOMContentLoaded', function () {
            const applyFiltersButton = document.getElementById('applyFilters');
            const nannyCatalog = document.getElementById('nannyCatalog');

            applyFiltersButton.addEventListener('click', function () {
              const experienceValue = parseInt(document.getElementById('filterExperience').value) || 0;
              const priceValue = parseFloat(document.getElementById('filterPrice').value) || Infinity;
              const ratingValue = parseFloat(document.getElementById('filterRating').value) || 0;

              // Проходим по всем карточкам нянь и фильтруем их
              const nannyCards = nannyCatalog.getElementsByClassName('nanny-card');
              Array.from(nannyCards).forEach(card => {
                const experience = parseInt(card.dataset.experience);
                const price = parseFloat(card.dataset.price);
                const rating = parseFloat(card.dataset.rating);

                if (experience >= experienceValue && price <= priceValue && rating >= ratingValue) {
                  card.style.display = 'block';
                } else {
                  card.style.display = 'none';
                }
              });
            });
          });
        `}</script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.1.3/js/bootstrap.bundle.min.js"></script>
      </body>
    </>
  );
}

export default Home;
