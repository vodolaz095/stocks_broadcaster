Stocks Broadcaster
=================================
Application subscribes to [GRPC stream](https://tinkoff.github.io/investAPI/marketdata/#marketdataserversidestream)
with last prices and broadcasts by pub/sub channels data via redis server to trade bots.

Create broker account for [T-Bank Open Investment API](https://www.tbank.ru/sl/AugaFvDlqEP)

Support development - https://www.tinkoff.ru/rm/ostroumov.anatoliy2/4HFzm76801/

Config
=================================
Configuration example
[stocks_broadcaster_example.yaml](contrib%2Fstocks_broadcaster_example.yaml)

Key meaning 
***input***

Define inputs' parameters - trade api token, local network address (can be omitted) and FIGI of instruments to subscribe.

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

Define parameters to render and route last price messages via redis pub/sub channels.

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

Define name and connection string for redis servers to broadcast last price updates
Message format - JSON in UTF8 encoding:

```json5
{
  "name": "tmos", // as defined in `name`
  "value": 5.73,  // price of lot / цена лота
  "error": "",    // free form error message / сообщение об ошибки
  "timestamp":"Sun Aug 25 2024 01:06:23 GMT+0300"
}
```

Message is published in channel defined in `channel` key of config.

Example:

```yaml
instruments:
  - figi: "BBG004730RP0"
    name: "GAZP"
    channel: "stocks/gazp"

```
will publish message 
```json5
{
  "name": "GAZP",
  "value": 5.73, 
  "error": "", 
  "timestamp":"Sun Aug 25 2024 01:06:23 GMT+0300"
}
```
into redis channel `stocks/gazp`.


***log***

Define logging parameters.

Development using golang compiler on host machine
=============================
Application requires modern linux machine (tested on fedora 39+) with [Golang 1.22.0](https://go.dev/dl/) and [GNU Make](https://www.gnu.org/software/make/) installed.

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

Redis can be started by docker/podman

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

