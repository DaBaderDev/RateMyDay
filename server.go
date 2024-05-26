package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	Id   int
	Done bool
	Task string
}

type TaskResponseData struct {
	Tasks []Task
}

func main() {
	db, err := sql.Open("sqlite3", "./db.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists tasks (id integer not null primary key, task text not null, done integer not null, date datetime not null)")
	if err != nil {
		log.Fatal(err)
		return
	}

	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("web/"))
	todoTmpl := template.Must(template.ParseFiles("templates/todo.html"))

	router.Handle("/", fs)

	router.HandleFunc("/api/date", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			currentDateTime := time.Now()
			fmt.Fprintf(w, "%d-%d-%d", currentDateTime.Year(), currentDateTime.Month(), currentDateTime.Day())
		}
	})

	router.HandleFunc("/api/tasks/today", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			data := TaskResponseData{
				Tasks: []Task{},
			}

			rows, err := db.Query("select * from tasks where date=date('now')")
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			for rows.Next() {
				var id int
				var task string
				var done bool
				var date time.Time

				err = rows.Scan(&id, &task, &done, &date)
				if err != nil {
					log.Fatal(err)
				}

				data.Tasks = append(data.Tasks, Task{Id: id, Task: task, Done: done})
			}

			todoTmpl.Execute(w, data)
		}
	})

	router.HandleFunc("/api/completion/{date}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		date := vars["date"]

		var taskCount int
		var tasksDone int

		err = db.QueryRow(fmt.Sprintf("select count(id) from tasks where date=('%s')", date)).Scan(&taskCount)
		if err != nil {
			log.Fatal(err)
		}
		err = db.QueryRow(fmt.Sprintf("select count(id) from tasks where date=('%s') and done=1", date)).Scan(&tasksDone)
		if err != nil {
			log.Fatal(err)
		}

		score := math.Round(float64(tasksDone) / float64(taskCount) * 100)

		fmt.Fprint(w, score)
	})

	http.ListenAndServe(":8080", router)
}
