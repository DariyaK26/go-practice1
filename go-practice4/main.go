package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Email   string  `db:"email"`
	Balance float64 `db:"balance"`
}

func main() {

	dsn := "postgres://user:password@localhost:5430/mydatabase?sslmode=disable"

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		log.Fatal("open db:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("ping db:", err)
	}

	fmt.Println("✅ Connected to DB")

	aliceID, err := UpsertUser(db, User{Name: "Alice", Email: "alice@example.com", Balance: 100})
	if err != nil {
		log.Fatal("UpsertUser(alice):", err)
	}
	bobID, err := UpsertUser(db, User{Name: "Bob", Email: "bob@example.com", Balance: 50})
	if err != nil {
		log.Fatal("UpsertUser(bob):", err)
	}

	fmt.Println("После upsert:")
	printUsers(db)

	if err := TransferBalance(db, aliceID, bobID, 30); err != nil {
		log.Fatal("TransferBalance:", err)
	}
	fmt.Printf("\nПосле TransferBalance(%d -> %d, 30):\n", aliceID, bobID)
	printUsers(db)

	u, err := GetUserByID(db, aliceID)
	if err != nil {
		log.Fatal("GetUserByID:", err)
	}
	fmt.Printf("\nПользователь id=%d: %+v\n", aliceID, u)
}

func UpsertUser(db *sqlx.DB, u User) (int, error) {
	const q = `
INSERT INTO users (name, email, balance)
VALUES (:name, :email, :balance)
ON CONFLICT (email)
DO UPDATE SET
  name = EXCLUDED.name
RETURNING id;`

	var id int
	rows, err := db.NamedQuery(q, u)
	if err != nil {
		return 0, fmt.Errorf("upsert user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("scan returning id: %w", err)
		}
	}
	return id, rows.Err()
}

func GetAllUsers(db *sqlx.DB) ([]User, error) {
	var users []User
	err := db.Select(&users, `SELECT id, name, email, balance FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	return users, nil
}

func GetUserByID(db *sqlx.DB, id int) (User, error) {
	var u User
	err := db.Get(&u, `SELECT id, name, email, balance FROM users WHERE id=$1`, id)
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func TransferBalance(db *sqlx.DB, fromID, toID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be > 0")
	}
	if fromID == toID {
		return errors.New("fromID and toID must be different")
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	minID, maxID := fromID, toID
	if minID > maxID {
		minID, maxID = toID, fromID
	}

	var users []User
	query := `SELECT id, name, email, balance FROM users WHERE id IN ($1, $2) FOR UPDATE`
	if err := tx.Select(&users, query, minID, maxID); err != nil {
		return fmt.Errorf("lock users: %w", err)
	}
	if len(users) != 2 {
		return errors.New("one or both users not found")
	}

	var fromUser, toUser *User
	for i := range users {
		if users[i].ID == fromID {
			fromUser = &users[i]
		} else if users[i].ID == toID {
			toUser = &users[i]
		}
	}
	if fromUser == nil || toUser == nil {
		return errors.New("from or to user missing")
	}

	if fromUser.Balance < amount {
		return fmt.Errorf("insufficient funds: have %.2f, need %.2f", fromUser.Balance, amount)
	}

	if _, err := tx.Exec(`UPDATE users SET balance = balance - $1 WHERE id = $2`, amount, fromID); err != nil {
		return fmt.Errorf("debit from user: %w", err)
	}
	if _, err := tx.Exec(`UPDATE users SET balance = balance + $1 WHERE id = $2`, amount, toID); err != nil {
		return fmt.Errorf("credit to user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func printUsers(db *sqlx.DB) {
	users, err := GetAllUsers(db)
	if err != nil {
		log.Println("printUsers:", err)
		return
	}
	for _, u := range users {
		fmt.Printf("id=%d | %-10s | %-20s | balance=%.2f\n", u.ID, u.Name, u.Email, u.Balance)
	}
}
