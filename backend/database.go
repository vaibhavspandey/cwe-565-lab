package main
import (
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"
)
var DB *sql.DB
func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./bank.db")
	if err != nil { log.Fatal(err) }
	createUsers := `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, password TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user');`
	createSessions := `CREATE TABLE IF NOT EXISTS sessions (session_token TEXT PRIMARY KEY, user_id INTEGER, FOREIGN KEY(user_id) REFERENCES users(id));`
	DB.Exec(createUsers)
	DB.Exec(createSessions)
	DB.Exec("DELETE FROM users")
	DB.Exec("INSERT INTO users (username, password, role) VALUES ('alice', 'password123', 'user')")
	DB.Exec("INSERT INTO users (username, password, role) VALUES ('admin', 'adminpass', 'admin')")
}
