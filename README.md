Stocks Broadcaster
=================================
Application subscribes to [GRPC stream](https://tinkoff.github.io/investAPI/marketdata/#marketdataserversidestream)
with last prices and broadcasts by pub/sub channels data via redis server to trade bots.

Приложение подписывается на GRPC-поток и транслирует котировки через каналы (pub/sub channels) базы данных redis для торговых ботов.

Create broker account / Открыть брокерский счёт в [T-Bank Open Investment API](https://www.tbank.ru/sl/AugaFvDlqEP)

Постановка задания
================================
На данный момент (10 сентября 2024 года) брокер Т-Инвестиций предоставляет доступ к торгам
9383 типов акций, 4745 облигаций, 2125 фондов, 38 валют и 1552 фьючерсов и на котировки этих инструментов
можно подписаться. Open Investement API обладает жёсткими [лимитами](https://russianinvestments.github.io/investAPI/limits/) - 
300 подписок на котировки с одного ключа (и, похоже, с одного IP адреса).

Как разработчик торговых ботов, я хочу получать котировки акций (цена крайней сделки) через каналы базы данных
redis. Для обхода лимита в 300 подписок я хочу использовать систему, которая может создавать исходящие соединения
с разных сетевых интерфейсов, подключенных к серверу. 
То есть, я могу, используя токен `secret1` получать котировки фондов через локальный сетевой интерфейс с 
адресом `192.168.12.2`, а котировки акций я могу получать через интерфейс с `192.168.12.3` по токену `SECRET2`.


```yaml

inputs:
  - name: "etfs"
    token: "<<<SECRET1>>>"
    figis:
      - "BBG333333333"
      - ...
    local_addr: "192.168.12.2"
  - name: "stocks"
    token: "<<<SECRET2>>>"
    figis:
      - "BBG004730RP0"
      - "BBG00475KKY8"
      - ...
    local_addr: "192.168.12.3"

```

Данный подход (ретрансляция котировок из нескольких GRPC подписок в каналы базы данных redis) позволяет 
обойти лимиты на 300 подписок с токена\IP адреса, упростить код торгового робота (не нужно тащить GRPC клиент),
хотя и вносит лаг около 20 мс - что, в принципе, терпимо для алготрейдинга.


Config
=================================
Configuration example / Образец конфигурации
[stocks_broadcaster_example.yaml](contrib%2Fstocks_broadcaster_example.yaml)

Key meaning / Значение ключей конфигурации

***input***
Define inputs' parameters - trade api token and FIGI of instruments to subscribe / Задать параметры ввода - токен подключения
к API и FIGI инструментов, на котировки которых нужно подписаться

***instruments***
Define parameters to render and route last price messages via redis pub/sub channels / 
Задаёт направление и формат сообщения котировок, которое будет посылаться в каналы редиса. 

```yaml

instruments:
  - figi: "BBG333333333"
    name: "tmos"
    channel: "stocks/tmos"
  - figi: "BBG004730RP0"
    name: "GAZP"
    channel: "stocks/gazp"
  - figi: "BBG00475KKY8"
    name: "NVTK"
    channel: "stocks/NVTK"

```

***outputs***
Define name and connection string for redis servers to broadcast last price updates /
задать название и строку соединения до сервера redis, куда будут передаваться котировки.


Message format - JSON in UTF8 encoding / формат сообщения JSON в кодировке UTF-8

```json5
{
  "name": "tmos", // as defined in `name`
  "value": 5.73,  // price of lot / цена лота
  "error": "",    // free form error message / сообщение об ошибки
  "timestamp":"Sun Aug 25 2024 01:06:23 GMT+0300"
}
```

Message is published in channel defined in `channel` key of config / ключ конфигурации `channel` задаёт название канала,
куда публикуется сообщение.


Example / Пример:

```yaml
instruments:
  - figi: "BBG004730RP0"
    name: "GAZP"
    channel: "stocks/gazp"

```
will publish message to / опубликует сообщение 
```json5
{
  "name": "GAZP",
  "value": 5.73, 
  "error": "", 
  "timestamp":"Sun Aug 25 2024 01:06:23 GMT+0300"
}
```
в канал `stocks/gazp`.


***log***
Define logging parameters / Задать параметры логирования
