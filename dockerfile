FROM golang:1.21-alpine as builder

WORKDIR /src

COPY . .

RUN go mod download

EXPOSE 8002

RUN go build -o /elka main.go

###### final

FROM alpine:20240329

ENV TZ=UTC

RUN apk add --no-cache tzdata
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY --from=builder /elka /app/elka

ADD templates /app/templates 
ADD static /app/static 

RUN chmod +x /app/elka

ENV TZ=UTC

EXPOSE 8002

CMD ["/app/elka"]