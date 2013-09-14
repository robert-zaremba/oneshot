package main

import (
	_ "code.google.com/p/gosqlite/sqlite3"
	"database/sql"
	"github.com/scale-it/go-log"
	"github.com/scale-it/go-web/handlers"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var logger = log.NewStd(os.Stderr, log.Levels.Trace, log.Ldate|log.Lmicroseconds|log.Lshortfile, true)
var db *sql.DB
var token string

func setupDB() {
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		logger.Fatal(err)
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS 'task' (
    'id' INTEGER PRIMARY KEY,
    'job' VARCHAR(40) REFERENCES job (id),
    'user' VARCHAR(40),
    'created' DATE,
    'fetched' DATE NULL
);`); err != nil {
		logger.Fatal(err)
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS 'job' (
   'id' VARCHAR(40) PRIMARY KEY,
   'source' VARCHAR(40)
);
`); err != nil {
		logger.Fatal(err)
	}

}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "db.sqlite")
	setupDB()
	if err != nil {
		logger.Fatal("Can't open DB", err)
	}
	logger.Info("DB 'db.sqlite' initialized")

	rend := func(h handlers.HandlerRend) http.Handler {
		return handlers.Renderer{logger, h}
	}
	rand.Seed(time.Now().UnixNano())
	http.Handle("/oneShot/assign", withAuth{rend(hNewTask)})
	http.Handle("/oneShot/newjob", withAuth{rend(hNewJob)})
	http.Handle("/oneShot/gettask", rend(hGetTask))
	logger.Info("listening at 8080")
	http.ListenAndServe(":8080", nil) //http.FileServer(http.Dir("/usr/share/doc")))
}
