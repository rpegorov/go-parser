Test project for [go-parser](https://github.com/rpegorov/go-parser) library.
Test optimised parse big data (ower 1 billion records) from dpa-server.


1 Потребление памяти среднее 300 мб. Оптимально для работы 1 гб с учетом просадок.
2 Для получения данных через интернет параллельно обрабатывается не более 5 запросов.
3 Таймаут соединения с АПИ ДПА должен быть менее 60 секунд. Оптимально 120-240.
