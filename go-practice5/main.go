package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Movie struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
	ActorCount int    `json:"actor_count"`
}

func main() {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/movies", moviesHandler(db))

	addr := ":8080"
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func moviesHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var (
			yearMin *int
			yearMax *int
			limit   = 50
			offset  = 0
		)

		if v := r.URL.Query().Get("year_min"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				yearMin = &n
			}
		}
		if v := r.URL.Query().Get("year_max"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				yearMax = &n
			}
		}
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				limit = n
			}
		}
		if v := r.URL.Query().Get("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}

		baseSQL := `
SELECT
    m.id,
    m.title,
    m.year,
    COUNT(a.id) AS actor_count
FROM movies AS m
LEFT JOIN actors AS a ON a.movie_id = m.id
WHERE 1=1
`
		args := []any{}
		if yearMin != nil {
			baseSQL += " AND m.year >= $1"
			args = append(args, *yearMin)
		}
		if yearMax != nil {
			placeholder := "$1"
			if len(args) == 1 {
				placeholder = "$2"
			} else if len(args) == 2 {
				placeholder = "$3"
			}
			baseSQL += " AND m.year <= " + placeholder
			args = append(args, *yearMax)
		}

		limitPH := "$1"
		offsetPH := "$2"
		switch len(args) {
		case 1:
			limitPH = "$2"
			offsetPH = "$3"
		case 2:
			limitPH = "$3"
			offsetPH = "$4"
		}

		baseSQL += `
GROUP BY m.id
ORDER BY m.year DESC
LIMIT ` + limitPH + ` OFFSET ` + offsetPH

		args = append(args, limit, offset)

		start := time.Now()
		rows, err := db.QueryContext(ctx, baseSQL, args...)
		if err != nil {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("query error: %v", err)
			return
		}
		defer rows.Close()

		var out []Movie
		for rows.Next() {
			var m Movie
			if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.ActorCount); err != nil {
				http.Error(w, "scan error", http.StatusInternalServerError)
				log.Printf("scan error: %v", err)
				return
			}
			out = append(out, m)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "rows error", http.StatusInternalServerError)
			log.Printf("rows error: %v", err)
			return
		}
		elapsed := time.Since(start)

		log.Printf("SQL took: %s", elapsed)
		w.Header().Set("X-Query-Time", elapsed.String())

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
			return
		}
	})
}
