Stocks Broadcaster
=================================

Приложение подписывается на [GRPC поток](https://tinkoff.github.io/investAPI/marketdata/#marketdataserversidestream)
котировок и ретранслирует данные через каналы (pub/sub channels) базы данных redis для торговых ботов.
Открыть брокерский счёт в [T-Bank Open Investment API](https://www.tbank.ru/sl/AugaFvDlqEP)
Поддержать разработчика - https://www.tbank.ru/rm/ostroumov.anatoliy2/4HFzm76801/

Постановка задания
================================
Тестовое задание, выполнено примерно за 4 дня.

На данный момент (10 сентября 2024 года) брокер Т-Инвестиций предоставляет доступ к торгам
9383 типов акций, 4745 облигаций, 2125 фондов, 38 валют и 1552 фьючерсов и на котировки этих инструментов
можно подписаться. Open Investment API обладает жёсткими [лимитами](https://russianinvestments.github.io/investAPI/limits/) - 
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

Также я хочу запускать приложение как systemd unit на Cents 9 / Fedora 40 Server.

В версии v1.1.0 добавлен HTTP сервер с эндпоинтами для проверки готовности и жизнеспособности 
сервиса платформой [Kubernets](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/),
а также эндпоинт `GET /metrics`, выдающий метрики процесса (и значения котировок) в формате [Prometheus Metrics Scapper](https://prometheus.io/docs/prometheus/latest/getting_started/#configure-prometheus-to-monitor-the-sample-targets)

В версии v1.2.0 в конфиг добавлен массив баз данных Victoria Metrics, куда можно отправлять котировки и метрики процесса.


Конфигурация
=================================
Образец конфигурации - [stocks_broadcaster_example.yaml](contrib%2Fstocks_broadcaster_example.yaml)
Значение ключей конфигурации:

***input***

Задать параметры ввода - токен подключения к API, локальный сетевой адрес (необязательно) и FIGI инструментов, на котировки которых нужно подписаться.

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

```

***instruments***

Задаёт название канала и формат генерации сообщения котировок, которое будет посылаться в каналы редиса. 

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

```yaml

outputs:
  - name: "container"
    redis_url: "redis://127.0.0.1:6379" # syntax - https://pkg.go.dev/github.com/redis/go-redis/v9#ParseURL

```

Задать название и строку соединения до сервера redis, куда будут передаваться котировки.
Формат сообщения JSON в кодировке UTF-8

```json5
{
  "name": "tmos", // as defined in `name`
  "value": 5.73,  // price of lot / цена лота
  "error": "",    // free form error message / сообщение об ошибки
  "timestamp":"Sun Aug 25 2024 01:06:23 GMT+0300"
}
```

Ключ конфигурации `channel` задаёт название канала, куда публикуется сообщение.

Пример:

```yaml
instruments:
  - figi: "BBG004730RP0"
    name: "GAZP"
    channel: "stocks/gazp"

```
опубликует сообщение 
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
Задать параметры логирования

Разработка с использованием компилятора и базы данных redis на хост машине.
=============================
Приложение протестированно на современном линуксе (tested on fedora 39+) с 
установленными пакетами из официальный репозиториев дистрибутива.
[Golang 1.22.0](https://go.dev/dl/) и [GNU Make](https://www.gnu.org/software/make/).

```shell

# install compiler tools
$ dnf install golang make redis 

# install redis database on host machine
$ dnf install -y redis && systemctl enable --now redis

# ensure development tools in place
$ make tools

# ensure golang modules are installed
$ make deps

# start application for development using configuration from contrib/local.yaml
$ make start

# build production grade binary at `build/stocks_broadcaster`
$ make build

```

База данных Redis может быть запущена через docker/podman.

```shell

# start development redis database  
$ make docker/resource
$ make podman/resource

```

Development using docker + docker compose
=============================
[GNU Make](https://www.gnu.org/software/make/), [Docker engine](https://docs.docker.com/engine/install/) with
[compose plugin](https://docs.docker.com/compose/install/linux/) should be installed.
Installing golang toolchain on host machine is not required.

```shell

# start development databases and build and start application on http://localhost:3001 
$ make docker/up

# start development databases  
$ make docker/resource

# stop all
$ make docker/down

# prune all development environment
$ make docker/prune


```


Development using podman + podman-compose
=============================
Installing golang toolchain on host machine is not required.
Tested on Fedora 39, 40 and Centos 9 Stream.

```shell

# install development environment
$ sudo dnf install make podman podman-compose podman-plugins containernetworking-plugins

# start development databases and build and start application on http://localhost:3001
$ make podman/up

# start development databases  
$ make podman/resource

# stop all
$ make podman/down

# prune all development environment
$ make podman/prune

```

Как проверить работоспособность приложения?
=============================

1. Задать параметры конфигурации, допустим, в `contrib/local.yaml`
2. Запустить базу данных и приложение каким-либо образом (`make start`, `make docker/up` и т.д.) 
3. Подключится к серверу `redis` с помощью `redis-cli` и вызвать команду `monitor`
4. Смотреть вывод консоли редиса - будет показаны вывозы команды `publish` от приложения на публикацию котировок в каналы.
   В частности, если сейчас идут торги акциями "Газпрома", то сообщения будут такими:

```

vodolaz095@steel:~$ redis-cli monitor
OK
1731687168.592596 [0 127.0.0.1:39872] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.25,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:48.397151Z\"}"
1731687168.592655 [0 127.0.0.1:39886] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.25,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:48.397151Z\"}"
1731687169.099057 [0 127.0.0.1:39886] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.27,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:48.882235Z\"}"
1731687169.099103 [0 127.0.0.1:39872] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.27,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:48.882235Z\"}"
1731687170.635182 [0 127.0.0.1:39886] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.28,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:50.443741Z\"}"
1731687170.635268 [0 127.0.0.1:39872] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.28,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:50.443741Z\"}"
1731687170.635787 [0 127.0.0.1:39872] "publish" "stocks/NVTK" "{\"name\":\"NVTK\",\"value\":919.6,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:50.466195Z\"}"
1731687170.635985 [0 127.0.0.1:39886] "publish" "stocks/NVTK" "{\"name\":\"NVTK\",\"value\":919.6,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:50.466195Z\"}"
1731687171.352578 [0 127.0.0.1:39872] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.26,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:51.179342Z\"}"
1731687171.352624 [0 127.0.0.1:39886] "publish" "stocks/gazp" "{\"name\":\"GAZP\",\"value\":133.26,\"error\":\"\",\"timestamp\":\"2024-11-15T16:12:51.179342Z\"}"
^C

```
5. На уровне логгирования `debug` приложение будет писать примерно такой вывод:
```

vodolaz095@steel:~/projects/stocks_broadcaster$ make start 
go run main.go ./contrib/local.yaml
19:15:40 INF main.go:61 > Starting StockBroadcaster version development. GOOS: linux. ARCH: amd64. Go Version: go1.22.8. Please, report bugs here: https://github.com/vodolaz095/stocks_broadcaster/issues
19:15:40 INF main.go:77 > Reader etfs uses local address 192.168.47.9 to dial invest API
19:15:40 WRN main.go:150 > Systemd watchdog disabled - application can work unstable in systemd environment
19:15:40 DBG service_start.go:19 > Preparing to start reader 0 InvestAPI reader etfs...
19:15:40 DBG service_start.go:19 > Preparing to start reader 1 InvestAPI reader stocks...
19:15:40 DBG service_subscribe.go:16 > Creating subscription channel for writer 0 container1...
19:15:40 DBG service_subscribe.go:16 > Creating subscription channel for writer 1 container2...
19:15:40 TRC reader.go:97 > Reader InvestAPI reader stocks received: subscribe_last_price_response:{tracking_id:"61ac060fc2703aef8e3c7b48a6c40ab9" last_price_subscriptions:{figi:"BBG004730RP0" subscription_status:SUBSCRIPTION_STATUS_SUCCESS instrument_uid:"962e2a95-02a9-4171-abd7-aa198dbe643a" stream_id:"88a48d6e-7896-47cc-99ce-d8a3aff381b0" subscription_id:"5c50ec81-79ff-4d08-aafb-971d9cc51edf"} last_price_subscriptions:{figi:"BBG00475KKY8" subscription_status:SUBSCRIPTION_STATUS_SUCCESS instrument_uid:"0da66728-6c30-44c4-9264-df8fac2467ee" stream_id:"88a48d6e-7896-47cc-99ce-d8a3aff381b0" subscription_id:"303860b0-b5a6-41b0-b5a3-35fd02a92fce"}}
19:15:40 TRC reader.go:97 > Reader InvestAPI reader etfs received: subscribe_last_price_response:{tracking_id:"5527ca8063a72475382a677df49faf22" last_price_subscriptions:{figi:"BBG333333333" subscription_status:SUBSCRIPTION_STATUS_SUCCESS instrument_uid:"9654c2dd-6993-427e-80fa-04e80a1cf4da" stream_id:"7165bfb2-e045-47c3-aeec-23c2cbfa00c1" subscription_id:"5078900a-d949-4221-9bc3-0f2ddc9e0186"}}
19:15:41 TRC reader.go:97 > Reader InvestAPI reader stocks received: last_price:{figi:"BBG004730RP0" price:{units:133 nano:490000000} time:{seconds:1731687340 nanos:971043000} instrument_uid:"962e2a95-02a9-4171-abd7-aa198dbe643a" last_price_type:LAST_PRICE_EXCHANGE}
19:15:41 DBG reader.go:100 > Reader InvestAPI reader stocks: instrument BBG004730RP0 has last lot price 133.4900
19:15:43 TRC reader.go:97 > Reader InvestAPI reader stocks received: last_price:{figi:"BBG00475KKY8" price:{units:921 nano:800000000} time:{seconds:1731687343 nanos:433960000} instrument_uid:"0da66728-6c30-44c4-9264-df8fac2467ee" last_price_type:LAST_PRICE_EXCHANGE}
19:15:43 DBG reader.go:100 > Reader InvestAPI reader stocks: instrument BBG00475KKY8 has last lot price 921.8000
19:15:43 TRC reader.go:97 > Reader InvestAPI reader stocks received: last_price:{figi:"BBG00475KKY8" price:{units:922} time:{seconds:1731687343 nanos:611561000} instrument_uid:"0da66728-6c30-44c4-9264-df8fac2467ee" last_price_type:LAST_PRICE_EXCHANGE}
19:15:43 DBG reader.go:100 > Reader InvestAPI reader stocks: instrument BBG00475KKY8 has last lot price 922.0000
19:15:43 TRC reader.go:97 > Reader InvestAPI reader stocks received: last_price:{figi:"BBG004730RP0" price:{units:133 nano:500000000} time:{seconds:1731687343 nanos:586901000} instrument_uid:"962e2a95-02a9-4171-abd7-aa198dbe643a" last_price_type:LAST_PRICE_EXCHANGE}
19:15:43 DBG reader.go:100 > Reader InvestAPI reader stocks: instrument BBG004730RP0 has last lot price 133.5000
19:15:44 TRC reader.go:97 > Reader InvestAPI reader stocks received: last_price:{figi:"BBG00475KKY8" price:{units:921 nano:800000000} time:{seconds:1731687344 nanos:567429000} instrument_uid:"0da66728-6c30-44c4-9264-df8fac2467ee" last_price_type:LAST_PRICE_EXCHANGE}
19:15:44 DBG reader.go:100 > Reader InvestAPI reader stocks: instrument BBG00475KKY8 has last lot price 921.8000
19:15:45 TRC reader.go:97 > Reader InvestAPI reader stocks received: last_price:{figi:"BBG00475KKY8" price:{units:922} time:{seconds:1731687345 nanos:742633000} instrument_uid:"0da66728-6c30-44c4-9264-df8fac2467ee" last_price_type:LAST_PRICE_EXCHANGE}
19:15:45 DBG reader.go:100 > Reader InvestAPI reader stocks: instrument BBG00475KKY8 has last lot price 922.0000
19:15:45 TRC reader.go:97 > Reader InvestAPI reader etfs received: last_price:{figi:"BBG333333333" price:{units:5 nano:880000000} time:{seconds:1731687345 nanos:793230000} instrument_uid:"9654c2dd-6993-427e-80fa-04e80a1cf4da" last_price_type:LAST_PRICE_EXCHANGE}
19:15:45 DBG reader.go:100 > Reader InvestAPI reader etfs: instrument BBG333333333 has last lot price 5.8800
^C19:15:46 INF main.go:172 > Signal interrupt is received
19:15:46 INF reader.go:72 > Reader InvestAPI reader stocks is closing grpc subscription for 2 instruments
19:15:46 DBG reader.go:90 > Connection for InvestAPI reader stocks is canceled
19:15:46 DBG service_subscribe.go:24 > Closing subscription channel for writer 1 container2...
19:15:46 DBG service_start.go:91 > Closing broadcasting...
19:15:46 DBG service_subscribe.go:27 > Subscription channel for writer 1 container2 is closed
19:15:46 DBG reader.go:90 > Connection for InvestAPI reader etfs is canceled
19:15:46 DBG service_subscribe.go:24 > Closing subscription channel for writer 0 container1...
19:15:46 DBG service_subscribe.go:27 > Subscription channel for writer 0 container1 is closed
19:15:46 DBG service_start.go:34 > 2 readers are closing
19:15:46 INF reader.go:72 > Reader InvestAPI reader etfs is closing grpc subscription for 1 instruments
19:15:46 DBG service_close.go:18 > close: reader InvestAPI reader etfs is terminated
19:15:46 DBG service_close.go:18 > close: reader InvestAPI reader stocks is terminated
19:15:46 DBG service_close.go:28 > close: writer container1 is terminated
19:15:46 DBG service_close.go:28 > close: writer container2 is terminated
19:15:46 INF service_close.go:30 > close: system is stopped
19:15:46 INF main.go:196 > Stocks Broadcaster is terminated.
make: *** [Makefile:38: start] Ошибка 1

```
