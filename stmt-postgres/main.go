package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	_ "github.com/lib/pq"
)

type Storage struct {
	conn        *sql.DB
	getUserStmt *sql.Stmt
}

const DbName = "mydatabase"
const DbUser = "postgres"
const DbPass = "mysecretpassword"

/*
dbURL might look like:
"postgres://username:password@localhost:5432/database_name"
*/

type UserRec struct {
	User        string `json: "user"`
	DisplayName string `json: "display_name"`
	ID          int    `json: "id"`
}

func NewStorage(ctx context.Context, conn *sql.DB) *Storage {
	stmt, err := conn.PrepareContext(ctx, `SELECT "display_name", "user" FROM users WHERE id = $1`)
	if err != nil {
		log.Fatal(err)
	}
	return &Storage{
		getUserStmt: stmt,
	}
}

func (s *Storage) GetUser(ctx context.Context, id int) (UserRec, error) {
	u := UserRec{ID: id}
	err := s.getUserStmt.QueryRow(id).Scan(&u.User, &u.DisplayName)
	return u, err
}

func AddorUpdateUser(ctx context.Context, s *sql.DB, u []UserRec) (err error) {
	const (
		getStmt    = `SELECT "id" FROM users WHERE "user" = $1`
		insertStmt = `INSERT INTO users ("user",display_name) VALUES ($1, $2)`
		updateStmt = `UPDATE "users" SET "user" = $1, "display_name" = $2 WHERE "id" = $3`
	)
	tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	for _, user := range u {
		row := tx.QueryRowContext(ctx, getStmt, user.User)
		if err := row.Scan(&user.ID); err != nil {
			if err == sql.ErrNoRows {
				_, err = tx.ExecContext(ctx, insertStmt, user.User, user.DisplayName)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		fmt.Println("Here: ", &user)
		_, err = tx.ExecContext(ctx, updateStmt, user.User, user.DisplayName, user.ID)
		if err != nil {
			return err
		}
		// _, err = tx.ExecContext(context.Background(), insertStmt, user.User, user.DisplayName)
		// if err != nil {
		// 	return fmt.Errorf("tx.ExecContext %w", err)
		// }
	}
	return nil
}

func main() {
	// connectionString := fmt.Sprintf("postgres://%v:%v@127.0.0.1:5432/%v?sslmode=disable", DbUser, DbPass, DbName)
	connectionString := url.URL{
		Scheme: "postgres",
		Host:   "127.0.0.1:5432",
		User:   url.UserPassword(DbUser, DbPass),
		Path:   DbName,
	}
	q := connectionString.Query()
	q.Add("sslmode", "disable")
	connectionString.RawQuery = q.Encode()

	conn, err := sql.Open("pgx", connectionString.String())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := conn.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	// // // var u UserRec
	u := []UserRec{
		{
			User:        "udele",
			DisplayName: "udi",
			ID:          13,
		},
		{
			User:        "thankgod",
			DisplayName: "thank",
			ID:          14,
		},
		{
			User:        "ukpana",
			DisplayName: "ukpa",
			ID:          15,
		},
	}

	if err = AddorUpdateUser(ctx, conn, u); err != nil {
		fmt.Println("Insert", err)
	}
	fmt.Println("Executed succesfully")

	defer cancel()
}
