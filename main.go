package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//Time type stryct
type date_time struct {
	Time string `json: "time"`
}

// User type struct
type User struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Name    string             `bson:"name",omitempty`
	Surname string             `bson:"surname",omitempty`
	Dob     string             `bson:"dob",omitempty`
}

// Init Users varibale
// This variables stoors user data in memory
var Users []User

// Database connection struct
type Connection struct {
	Users *mongo.Collection
}

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

func (connection Connection) getUsers(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")
	var users []User

	//Get parameters value
	params := req.URL.Query()
	// If no parameters Encode all users

	if len(params) == 0 {
		// create a filter (cursor)
		// this filter returns all entries in the database
		cursor, err := connection.Users.Find(context.TODO(), bson.M{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `" }`))
		}
		//Apply filter to all entries
		cursor.All(context.TODO(), &users)

		// repond with filtered content
		json.NewEncoder(w).Encode(users)
		return
	}

	//Go through all names
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

func (connection Connection) createUsers(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Create new user var and decode json contect from body
	var user User
	_ = json.NewDecoder(req.Body).Decode(&user)

	// insert user into database
	result, _ := connection.Users.InsertOne(context.TODO(), user)
	//Response with json data
	json.NewEncoder(w).Encode(result)
}

func (connection Connection) getUser(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")
	// retrieve map of veriables from get url
	param := mux.Vars(req)

	// Gat object ID
	objectId, err := primitive.ObjectIDFromHex(param["id"])
	if err != nil {
		log.Println("Invalid ID")
	}

	// Find document with sepcified ID
	var user User
	//TODO: Do not show ID field
	connection.Users.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&user)

	// repond with user
	json.NewEncoder(w).Encode(user)
}

func (connection Connection) updateUser(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")
	// retrieve map of veriables from get url
	param := mux.Vars(req)
	// Gat object ID
	objectId, err := primitive.ObjectIDFromHex(param["id"])
	if err != nil {
		log.Println("Invalid ID")
	}
	var user User
	// decode json in request body
	json.NewDecoder(req.Body).Decode(&user)
	// update specified user
	result, err := connection.Users.UpdateOne(
		context.TODO(),          // required context
		bson.M{"_id": objectId}, // filter
		bson.D{{"$set", user}},
	)

	json.NewEncoder(w).Encode(result)

}

func (connection Connection) deleteUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Create new user var and decode json contect from body
	param := mux.Vars(req)
	//Get Object id
	objectId, err := primitive.ObjectIDFromHex(param["id"])
	if err != nil {
		log.Println("Invalid ID")
	}
	result, err := connection.Users.DeleteOne(context.TODO(), bson.M{"_id": objectId})

	//Response with json data
	json.NewEncoder(w).Encode(result)

}

func main() {
	// connect to mongodb
	opts := options.Client().ApplyURI("mongodb://mongodb:27017")
	clientdb, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	//ping db
	err = clientdb.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected and pinged mongodb")

	collectionUsers := clientdb.Database("myDB").Collection("Users")
	connection := Connection{
		Users: collectionUsers,
	}

	// init server mux
	router := mux.NewRouter()

	//Handelers
	router.HandleFunc("/time", getTime).Methods("GET")
	router.HandleFunc("/users", connection.getUsers).Methods("GET")
	router.HandleFunc("/users", connection.createUsers).Methods("POST")
	router.HandleFunc("/users/{id}", connection.getUser).Methods("GET")
	router.HandleFunc("/users/{id}", connection.updateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", connection.deleteUser).Methods("DELETE")

	// listen and serve requests on localhost port 8080
	// Use server mux router
	log.Fatal(http.ListenAndServe(":8081", router))

}
