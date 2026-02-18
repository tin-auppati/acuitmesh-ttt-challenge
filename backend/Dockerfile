# backend/Dockerfile
FROM golang:1.22-alpine

WORKDIR /app

# Install air for live reload (Optional but recommended for dev)
RUN go install github.com/cosmtrek/air@latest

COPY go.mod ./

RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]
