package main

import (
	"encoding/json"
	"io"
	"net/http"

	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"

	"github.com/rs/cors"
)

var db, _ = gorm.Open("mysql", "root:root@tcp(127.0.0.1:3307)/todolist?charset=utf8&parseTime=True&loc=Local")

//TodoItemModel struct
type TodoItemModel struct {
	ID          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {
	defer db.Close()

	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	log.Info("Starting todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo-completed", getCompletedItems).Methods("GET")
	router.HandleFunc("/todo-incomplete", getIncompleteItems).Methods("GET")
	router.HandleFunc("/todo", createItem).Methods("POST")
	router.HandleFunc("/todo/{id}", updateItem).Methods("PUT")
	router.HandleFunc("/todo/{id}", deleteItem).Methods("DELETE")
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)
	http.ListenAndServe(":8085", handler)
}

//Healthz function
func Healthz(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is good")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue(("description"))
	log.WithFields(log.Fields{"description": description}).Info("Add new ToDoItem. Saving to database")
	todo := &TodoItemModel{Description: description, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Value)
}

func getItemByID(ID int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, ID)
	if result.Error != nil {
		log.Warn("Not Found!")
		return false
	}
	return true
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Check Exist Data
	err := getItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"update": false, "error": "Not Found"}`)
	} else {
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating ToDo")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		todo.Completed = completed
		db.Save(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Check Exist Data
	err := getItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"update": false, "error": "Not Found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Delete ToDo")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func getTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func getCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get completed TodoItems")
	completedTodoItems := getTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

func getIncompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get incomplete TodoItems")
	incompleteTodoItems := getTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incompleteTodoItems)
}
