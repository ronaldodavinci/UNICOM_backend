# ---- build stage ----
FROM golang:1.24.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# ---- runtime stage ----
FROM alpine:latest

RUN apk add --no-cache tzdata

WORKDIR /root/

COPY --from=build /app/api .

EXPOSE 8000

CMD ["./api"]