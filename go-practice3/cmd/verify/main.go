package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// fallback: укажи твой DSN тут, если не хочешь через env
		dsn = "postgres://exp_user:verysecret@localhost:5432/expense_tracker?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	// Проверим, что таблицы есть
	var count int
	q := `
SELECT COUNT(*) 
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name IN ('users','categories','expenses');`
	if err := db.QueryRow(q).Scan(&count); err != nil {
		log.Fatalf("query: %v", err)
	}

	fmt.Printf("DB OK. Found %d of 3 expected tables.\n", count)
}
