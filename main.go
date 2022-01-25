package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

//Time type stryct
type date_time struct {
	Time string `json: "time"`
}

// User type struct
type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

// Init Users varibale
// This variables stoors user data in memory
var Users []User

//Handlers
func getTime(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")

	// Get and format default time responce
	current_time := time.Now()
	t := current_time.Format("2006-01-02 15:04:05")

	//Check Query paraments for additional formating
	param_time := req.URL.Query().Get("time_only")
	param_data := req.URL.Query().Get("date_only")
	if param_time != "" {
		t = current_time.Format("15:04:05")
	}
	if param_data != "" {
		t = current_time.Format("2006-01-02")
	}

	//Time veriable of struct date_time
	time := date_time{
		Time: t,
	}

	// Encode time as json data
	json.NewEncoder(w).Encode(time)
}

func getUsers(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")

	//Get parameters value
	params := req.URL.Query()
	if params == nil {
		//Encode all users
		json.NewEncoder(w).Encode(Users)
		return
	}

	//Go thouch all names
	searchName := req.URL.Query().Get("name")
	var selectUsers []User
	for _, item := range Users {
		if strings.Contains(item.Name, searchName) {
			selectUsers = append(selectUsers, item)
		}
	}

	//Encode all users
	json.NewEncoder(w).Encode(selectUsers)

}

func createUsers(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Create new user var and decode json contect in body
	var user User
	_ = json.NewDecoder(req.Body).Decode(&user)
	//Give uuid to new user
	user.ID = uuid.NewString()
	//Add user to memory
	Users = append(Users, user)
	//Response with json data
	json.NewEncoder(w).Encode(user)
}

func getUser(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")
	// retrieve map of veriables from get url
	param := mux.Vars(req)
	// loop through books to find id
	for _, item := range Users {
		if item.ID == param["id"] {
			// Encode books as json data
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	// catch all
	// encode empty user
	json.NewEncoder(w).Encode(&User{})
}

func updateUser(w http.ResponseWriter, req *http.Request) {
	//Get user id from url
	path := req.URL.Path
	split := strings.Split(path, "/")
	searchId := split[len(split)-1]
	//loops through data and updates specified user
	for i, item := range Users {
		if item.ID == searchId {
			var tempUser User
			//decode json data and override user detail
			_ = json.NewDecoder(req.Body).Decode(&tempUser)
			Users[i].Name = tempUser.Name
			Users[i].Surname = tempUser.Surname
			json.NewEncoder(w).Encode(Users[i])
		}
	}
}

func deleteUser(w http.ResponseWriter, req *http.Request) {
	//Get user id from url
	path := req.URL.Path
	split := strings.Split(path, "/")
	searchId := split[len(split)-1]
	var tempUsers []User
	//creates new slice without deleted user !very inefficent
	for _, item := range Users {
		if item.ID != searchId {
			tempUsers = append(tempUsers, item)
		}
	}
	//copy temp users over to Users
	Users = tempUsers
}

func main() {
	// init server mux
	router := mux.NewRouter()

	//Mock Data
	Users = append(Users, User{ID: uuid.NewString(), Name: "Filip", Surname: "Van der Zel"})
	Users = append(Users, User{ID: uuid.NewString(), Name: "Jan", Surname: "Semmelink"})

	//Handelers
	router.HandleFunc("/time", getTime).Methods("GET")
	router.HandleFunc("/users", getUsers).Methods("GET")
	router.HandleFunc("/users", createUsers).Methods("POST")
	router.HandleFunc("/users/{id}", getUser).Methods("GET")
	router.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	// listen and serve requests on localhost port 8080
	// Use server mux router
	log.Fatal(http.ListenAndServe(":8081", router))

}
