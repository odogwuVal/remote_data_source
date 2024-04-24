package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

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

func GetUser(ctx context.Context, conn *sql.DB, id int) (UserRec, error) {
	const query = `SELECT "user","display_name" FROM users
	WHERE "id" = $1`
	u := UserRec{ID: id}
	err := conn.QueryRowContext(ctx, query, id).Scan(&u.User, &u.DisplayName)
	return u, err
}

func addUser(ctx context.Context, conn *sql.DB, u UserRec) (UserRec, error) {
	const query = `INSERT INTO users ("user",display_name) VALUES ($1, $2)`
	result, err := conn.ExecContext(ctx, query, u.User, u.DisplayName)
	if err != nil {
		return u, err
	}
	id, err := result.RowsAffected()
	if err != nil {
		return u, err
	}
	u.ID = int(id)
	return u, nil
}

func main() {
	connectionString := fmt.Sprintf("postgres://%v:%v@127.0.0.1:5432/%v?sslmode=disable", DbUser, DbPass, DbName)
	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	u := UserRec{
		User:        "ekene",
		DisplayName: "eky",
	}

	defer conn.Close()
	if err = conn.Ping(); err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	rec, err := GetUser(ctx, conn, 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rec)
	rec, err = addUser(ctx, conn, u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rec)
}
