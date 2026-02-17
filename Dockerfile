FROM golang:latest AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /sportify ./cmd/api

FROM alpine:latest AS run

WORKDIR /app

COPY --from=build /sportify /app/sportify
COPY ./cmd/api/static/* /app/static/
COPY ./cmd/api/index.html /app/

EXPOSE 8080

CMD ["/app/sportify"]
