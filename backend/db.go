// backend/db.go

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	var err error
	dsn := os.Getenv("DB_URL")

	if dsn == "" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)
	}

	// ลอง Connect (Retry ได้เผื่อ DB ยังไม่ตื่น)
	for i := 0; i < 5; i++ {
		DB, err = sql.Open("postgres", dsn)
		if err == nil {
			err = DB.Ping() // เช็คว่าต่อติดจริงๆ
		}

		if err == nil {
			fmt.Println("Connected to Database successfully!")
			return
		}

		fmt.Printf("Failed to connect to DB (Attempt %d/5). Retrying in 2s...\n", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("Could not connect to database after 5 attempts:", err)
}
