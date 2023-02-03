package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

//For localhost usage
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func main() {
	dbSource := fmt.Sprintf(
		"user=%s password=%s host=%s dbname=%s sslmode=disable port=%s",
		"dev",
		"dev",
		"localhost",
		"dev",
		"15432",
	)
	db = sqlx.MustOpen("postgres", dbSource)
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		panic("Unable to establish database connection: " + err.Error())
	}

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/add/employees", addEmployee)
			r.Post("/update/employees/{column}", updateEmployee)
			r.Get("/get/employees", getEmployees)
			r.Post("/get/employees/{column}", getEmployee)
			r.Get("/join/employees", joinEmployees)
			r.Post("/delete/employees/{column}", deleteEmployee)
		})
	})

	fmt.Println("Listening on port 1234")

	err := http.ListenAndServe(":1234", r)
	if err != nil {
		panic(err)
	}
}

//http://localhost:1234/v1/add/employees
func addEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	ctx := r.Context()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	newEmployee := &employee{}
	err = json.Unmarshal(body, &newEmployee)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if newEmployee.First_name == "" {
		http.Error(w, "Missing name", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	sqlStatement := "insert into employee(id, first_name) " +
		"values($1,$2)"

	_, err = db.ExecContext(ctx, sqlStatement, id, newEmployee.First_name)
	if err != nil {
		fmt.Printf("Insert failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	newEmployee.Id = id
	response, err := json.Marshal(newEmployee)
	if err != nil {
		fmt.Printf("Marshal failed failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(response)
}

//http://localhost:1234/v1/update/employees/{column}
func updateEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	ctx := r.Context()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	updateEmployee := &update{}
	err = json.Unmarshal(body, &updateEmployee)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if updateEmployee.Id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	column := chi.URLParam(r, "column")

	sqlStatement := fmt.Sprintf("update employee set %v = $1 where employee.id = $2", column)

	_, err = db.ExecContext(ctx, sqlStatement, updateEmployee.Value, updateEmployee.Id)
	if err != nil {
		fmt.Printf("Insert failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(updateEmployee)
	if err != nil {
		fmt.Printf("Marshal failed failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(response)
}

//http://localhost:1234/v1/join/employees
func joinEmployees(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	ctx := r.Context()

	sqlStatement := "select *, d.department_id from employee e join department d on d.department_id = e.department_id"

	var employees []*employee
	if err := db.SelectContext(
		ctx,
		&employees,
		sqlStatement); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Read failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(employees)
	if err != nil {
		fmt.Printf("Marshal failed failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(response)
}

//http://localhost:1234/v1/get/employees/{column}
func getEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	ctx := r.Context()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	getEmployee := &get{}
	err = json.Unmarshal(body, &getEmployee)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if getEmployee.Value == "" {
		http.Error(w, "Missing value", http.StatusBadRequest)
		return
	}

	column := chi.URLParam(r, "column")

	sqlStatement := fmt.Sprintf("select * from employee where employee.%v = $1", column)

	var employees []*employee
	if err := db.SelectContext(
		ctx,
		&employees,
		sqlStatement,
		getEmployee.Value); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Read failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(employees)
	if err != nil {
		fmt.Printf("Marshal failed failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(response)
}

//http://localhost:1234/v1/get/employees
func getEmployees(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	ctx := r.Context()

	sqlStatement := "select * from employee"

	var employees []*employee
	if err := db.SelectContext(
		ctx,
		&employees,
		sqlStatement); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Read failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(employees)
	if err != nil {
		fmt.Printf("Marshal failed failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(response)
}

//http://localhost:1234/v1/delete/employees/{column}
func deleteEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	ctx := r.Context()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	deleteEmployee := &delete{}
	err = json.Unmarshal(body, &deleteEmployee)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if deleteEmployee.Value == "" {
		http.Error(w, "Missing value", http.StatusBadRequest)
		return
	}

	column := chi.URLParam(r, "column")

	sqlStatement := fmt.Sprintf("delete from employee where employee.%v = $1", column)

	_, err = db.ExecContext(ctx, sqlStatement, deleteEmployee.Value)
	if err != nil {
		fmt.Printf("Insert failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(deleteEmployee)
	if err != nil {
		fmt.Printf("Marshal failed failed: %v\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(response)
}

type employee struct {
	Id              string         `json:"id,omitempty" db:"id"`
	First_name      string         `json:"first_name,omitempty" db:"first_name"`
	Last_name       sql.NullString `json:"last_name,omitempty" db:"last_name"`
	Phone           sql.NullString `json:"phone,omitempty" db:"phone"`
	Email           sql.NullString `json:"email,omitempty" db:"email"`
	Birthdate       sql.NullString `json:"birthdate,omitempty" db:"birthdate"`
	Startdate       sql.NullString `json:"startdate,omitempty" db:"startdate"`
	Enddate         sql.NullString `json:"enddate,omitempty" db:"enddate"`
	Salary          sql.NullString `json:"salary,omitempty" db:"salary"`
	Boss_id         sql.NullString `json:"boss_id,omitempty" db:"boss_id"`
	Department_id   sql.NullString `json:"department_id,omitempty" db:"department_id"`
	Created         sql.NullString `json:"created,omitempty" db:"created"`
	Department_name sql.NullString `json:"department_name,omitempty" db:"department_name"`
}

type update struct {
	Id    string `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
}

type delete struct {
	Value string `json:"value,omitempty"`
}

type get struct {
	Value string `json:"value,omitempty"`
}
