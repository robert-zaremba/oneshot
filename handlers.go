package main

import (
	"database/sql/driver"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	reName = regexp.MustCompile(`^[A-Za-z.]+$`)
	reNum  = regexp.MustCompile(`^[0-9]+$`)
)

type withAuth struct {
	H http.Handler
}

func (this withAuth) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if x := r.Header.Get("X-Auth-Token"); x != token {
		http.Error(rw, "", http.StatusUnauthorized)
		return
	}
	this.H.ServeHTTP(rw, r)
}

func hNewTask(rw http.ResponseWriter, r *http.Request) (interface{}, int) {
	var out = make(map[string]string)
	user := r.FormValue("user")
	if user == "" || !reName.MatchString(user) {
		out["user"] = "wrong format: " + user
	}
	job := r.FormValue("job")
	if job == "" || !reName.MatchString(job) {
		out["job"] = "wrong format: " + job
	}
	if len(out) != 0 {
		return out, http.StatusBadRequest
	}

	id := rand.Int()
	t := time.Now()
	logger.Debugf("new task: (%v, %v, %v, %v)", id, job, user, t)
	result, err := db.Exec("INSERT INTO 'task' VALUES (?,?,?,?,?)", id, job, user, t, nil)
	if err != nil {
		logger.Error(err)
		return err, http.StatusBadRequest
	}
	logger.Info(result.LastInsertId())

	return id, http.StatusOK
}

func hNewJob(rw http.ResponseWriter, r *http.Request) (interface{}, int) {
	var out = make(map[string]string)
	name := r.FormValue("name")
	if name == "" || !reName.MatchString(name) {
		out["name"] = "wrong format: " + name
	}
	file, fileH, err := r.FormFile("file")
	if err != nil {
		out["file"] = "wrong file: " + err.Error()
	} else {
		defer file.Close()
	}
	if len(out) != 0 {
		return out, http.StatusBadRequest
	}

	// fcontent, err := ioutil.ReadAll(file)
	// ioutil.WriteFile(fileH.Filename, fcontent, 0666)
	f, err := os.OpenFile(fileH.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	defer f.Close()
	io.Copy(f, file)

	logger.Debugf("new job: (%v, %v)", name, fileH.Filename)
	result, err := db.Exec("INSERT INTO 'job' VALUES (?,?)", name, fileH.Filename)
	if err != nil {
		logger.Error(err)
		return err, http.StatusBadRequest
	}
	job_num, _ := result.LastInsertId()
	logger.Info("new #job:", job_num)

	return name, http.StatusOK
}

func hGetTask(rw http.ResponseWriter, r *http.Request) (interface{}, int) {
	task := r.FormValue("task")
	if task == "" || !reNum.MatchString(task) {
		return "wrong task: " + task, http.StatusBadRequest
	}

	row := db.QueryRow("SELECT job, user, created, fetched, source FROM task JOIN job ON (task.job=job.id) WHERE task.id=?", task)
	var job, user, source string
	var created time.Time
	var fetched NullTime
	if err := row.Scan(&job, &user, &created, &fetched, &source); err != nil {
		logger.Error("DB error", err)
		return err, http.StatusBadRequest
	}
	logger.Debug("getting task:", task, job, source, created, " BY", user)
	if fetched.Valid {
		return fmt.Sprintf("Task '%s' has been already taken at %v", job, fetched.Time), http.StatusOK
	}

	_, err := db.Exec("UPDATE task SET fetched=? WHERE id=?", time.Now(), task)
	if err != nil {
		logger.Error(err)
		return err, http.StatusInternalServerError
	}

	http.ServeFile(rw, r, source)
	return "", 0
}

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
