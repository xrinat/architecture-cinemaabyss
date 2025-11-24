# --- Этап сборки ---
FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o events-service

# --- Этап запуска ---
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/events-service .

ENV PORT=8082
EXPOSE 8082

ENTRYPOINT ["/app/events-service"]
