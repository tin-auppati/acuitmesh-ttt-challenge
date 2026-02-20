
# Stage 1: Build Stage (ใช้ image ที่มี Go compiler)

FROM golang:1.25-alpine AS builder

WORKDIR /app

# ก๊อปปี้ไฟล์ go.mod และ go.sum มาโหลด dependencies ก่อน (เพื่อทำ caching)
COPY go.mod go.sum ./
RUN go mod download

# ก๊อปปี้โค้ดทั้งหมดเข้ามา
COPY . .

# สั่ง Build โค้ด Go ให้กลายเป็นไฟล์ binary เดี่ยวๆ ชื่อ "main"
RUN CGO_ENABLED=0 GOOS=linux go build -o main .


# Stage 2: Run Stage (ใช้ image เปล่าๆ เล็กๆ เพื่อประหยัด RAM)

FROM alpine:latest

WORKDIR /app

# ก๊อปปี้แค่ไฟล์ "main" ที่ build เสร็จแล้วมาจาก Stage 1
COPY --from=builder /app/main .

# สั่งรันแอปพลิเคชันโดยตรง (ไม่ผ่าน Air)
CMD ["./main"]