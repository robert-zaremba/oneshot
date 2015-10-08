package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
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
	os.Stdout.WriteString("RUNNING")
	usage := "usage:\n\t app <port to listen> <authentication token: [A-Za-z.]{8,}>"
	if len(os.Args) != 3 || len(os.Args[2]) < 8 || !reName.Match([]byte(os.Args[2])) {
		println(usage)
		logger.Fatal("wrong command line arguments")
	}
	token = os.Args[2]
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
	http.Handle("/oneShot/assign/", withAuth{rend(hNewTask)})
	http.Handle("/oneShot/newjob/", withAuth{rend(hNewJob)})
	http.Handle("/oneShot/gettask/", rend(hGetTask))
	logger.Info("listening on " + os.Args[1])
	err = http.ListenAndServe(":"+os.Args[1], nil)
	logger.Error(err)
}

// go run app.go handlers.go 8000 NTKJNiutiubvRONf
// curl -i -H "X-Auth-Token: NTKJNiutiubvRONf" "http://localhost:8000/oneShot/newjob/?name=rajesh" -F file=@./rajesh.tgz
// curl -i -H "X-Auth-Token: NTKJNiutiubvRONf" "http://localhost:8000/oneShot/assign/?job=ansii&user=robert"
// curl -i "http://localhost:8000/oneShot/gettask/?task="
