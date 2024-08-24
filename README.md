Stocks Broadcaster
=================================
Application subscribes to [GRPC stream](https://tinkoff.github.io/investAPI/marketdata/#marketdataserversidestream)
with last prices and broadcasts data via redis server to trade bots.

Приложение подписывается на GRPC-поток и транслирует котировки через потоки базы данных redis для торговых ботов.

Create broker account / Открыть брокерский счёт в [T-Bank Open Investment API](https://www.tbank.ru/sl/AugaFvDlqEP)

Config
=================================
Configuration example / Образец конфигурации
[stocks_broadcaster_example.yaml](contrib%2Fstocks_broadcaster_example.yaml)

Key meaning / Значение ключей конфигурации

***input***
Define inputs' parameters - trade api token and FIGI of instruments to subscribe / Задать параметры ввода - токен подключения
к API и FIGI инструментов, на котировки которых нужно подписаться

***instruments***

***outputs***

***log***
Define logging level / Задать параметры логгирование
