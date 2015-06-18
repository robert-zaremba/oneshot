package main

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	reName = regexp.MustCompile(`^[A-Za-z. ]+$`)
	reNum  = regexp.MustCompile(`^[0-9]+$`)
)

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

	row := db.QueryRow("SELECT job, created, fetched, source FROM task JOIN job ON (task.job=job.id) WHERE task.id=?", task)
	var job, created, source string
	var fetched interface{}
	if err := row.Scan(&job, &created, &fetched, &source); err != nil {
		logger.Warn("DB error", err)
		return err, http.StatusBadRequest
	}
	logger.Debug("getting task:", task, job, created, fetched, source)
	if fetched != nil {
		return "Task '" + job + "' has been already taken at " + string(fetched.([]byte)), http.StatusOK
	}

	_, err := db.Exec("UPDATE task SET fetched=? WHERE id=?", time.Now(), task)
	if err != nil {
		logger.Error(err)
		return err, http.StatusInternalServerError
	}

	http.ServeFile(rw, r, source)
	return "", 0
}
