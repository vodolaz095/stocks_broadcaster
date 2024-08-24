# this dockerfile builds minimal production grade image (~40mb) with alpine linux
FROM mirror.gcr.io/golang:1.22.6 AS builder

RUN mkdir -p /app
WORKDIR /app

ADD ./go.sum /app/go.sum
ADD ./go.mod /app/go.mod
RUN go mod download
RUN go mod verify

ADD . /app/
RUN make build

FROM mirror.gcr.io/alpine:3.14 AS runtime
RUN apk update && apk add --no-cache ca-certificates
COPY --from=builder /app/build/stocks_broadcaster /bin/stocks_broadcaster
EXPOSE 3000
ENTRYPOINT ["/bin/stocks_broadcaster"]
