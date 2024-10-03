function Home({ userName, userID, days, busyDays }) {
  return (
    <>
      <head>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Календарь</title>
        <link
          rel="stylesheet"
          href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css"
        />
        <style>{`
          body {
              padding: 20px;
          }
          h1 {
              margin-bottom: 20px;
          }
          /* Стиль для календаря */
          .calendar {
              display: flex;
              flex-wrap: wrap;
              margin-top: 20px;
          }
          .day {
              border: 1px solid #ddd;
              padding: 10px;
              width: 14.28%; /* 7 дней в ряду */
              box-sizing: border-box;
              text-align: center;
              margin-bottom: 10px;
          }
          .busy {
              background-color: #f8d7da;
          }
          .free {
              background-color: #d4edda;
          }
        `}</style>
      </head>

      <body>
        {/* Верхнее меню */}
        <div className="d-flex flex-column flex-md-row align-items-center p-3 px-md-4 mb-3 bg-white border-bottom shadow-sm">
          <h5 className="my-0 mr-md-auto font-weight-normal">Nanny Market</h5>
          <nav className="my-2 my-md-0 mr-md-3">
            <a className="p-2 text-dark" href={`/main?id=${userID}`}>
              Главная
            </a>
            <a className="p-2 text-dark" href={`/catalog?id=${userID}`}>
              Каталог нянь
            </a>
            <a className="p-2 text-dark" href={`/calendar?id=${userID}`}>
              Календарь
            </a>
          </nav>
        </div>

        <h1>Календарь пользователя {userName}</h1>

        <div className="calendar">
          {days.map((day, index) => (
            <div
              key={index}
              className={`day ${busyDays.includes(day) ? "busy" : "free"}`}
            >
              <strong>День {day}</strong>
              <br />
              {busyDays.includes(day) ? (
                <span>Занят</span>
              ) : (
                <button className="btn btn-info">Добавить событие</button>
              )}
            </div>
          ))}
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
