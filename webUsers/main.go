package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//Time type stryct
type timeResponse struct {
	Time string `json:"time"`
}

// User type struct
type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Surname  string             `json:"surname,omitempty" bson:"surname,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Username string             `json:"username,omitempty" bson:"username,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
	Dob      string             `json:"dob,omitempty" bson:"dob,omitempty"`
}

// Init Users varibale
// This variables stoors user data in memory
var Users []User

// Database connection struct
type Connection struct {
	Users *mongo.Collection
}

func main() {
	// connect to mongodb
	log.Println("Connecting to mongodb ...")
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Pinging...")

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	collectionUsers := client.Database("myDB").Collection("Users")
	connection := Connection{
		Users: collectionUsers,
	}

	// init server mux
	router := mux.NewRouter()

	//Handelers
	router.HandleFunc("/time", getTime).Methods("GET")
	router.HandleFunc("/verifyUser", connection.verifyUser).Methods("POST")
	router.HandleFunc("/users", connection.getUsers).Methods("GET")
	router.HandleFunc("/users", connection.createUsers).Methods("POST")
	router.HandleFunc("/users/{id}", connection.getUser).Methods("GET")
	router.HandleFunc("/users/{id}", connection.updateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", connection.deleteUser).Methods("DELETE")

	// listen and serve requests on localhost port 8081
	// Use server mux router
	log.Fatal(http.ListenAndServe(":8081", router))

}

//Handlers
func getTime(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")

	// Get and format default time responce
	currentTime := time.Now()
	t := currentTime.Format("2006-01-02 15:04:05")

	//Check Query paraments for additional formating
	param_time := req.URL.Query().Get("time_only")
	param_data := req.URL.Query().Get("date_only")
	if param_time != "" {
		t = currentTime.Format("15:04:05")
	}
	if param_data != "" {
		t = currentTime.Format("2006-01-02")
	}

	//Time veriable of struct timeResponse
	time := timeResponse{
		Time: t,
	}

	// Encode time as json data
	json.NewEncoder(w).Encode(time)

}

func (connection Connection) verifyUser(w http.ResponseWriter, req *http.Request) {
	u, p, ok := req.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}

	// Get Users password
	var user User
	connection.Users.FindOne(context.TODO(), bson.M{"username": u}).Decode(&user)
	password := user.Password
	if p != password {
		fmt.Printf("Password provided is incorrect: %s\n", u)
		w.WriteHeader(401)
		return
	}
	w.WriteHeader(200)
	return

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
			return
		}
		//Apply filter to all entries
		cursor.All(context.TODO(), &users)

		// repond with filtered content
		json.NewEncoder(w).Encode(users)
		return
	}

	var filter bson.D
	//Go through all names
	searchName := params.Get("name")
	if searchName != "" {
		filter = bson.D{primitive.E{
			Key: "name", Value: primitive.Regex{Pattern: searchName, Options: "i"}}}
	}

	searchUser := params.Get("username")
	if searchUser != "" {
		filter = bson.D{primitive.E{
			Key: "username", Value: searchUser}}
	}

	cursor, err := connection.Users.Find(context.TODO(), filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `" }`))
		return
	}

	//Apply filter to all entries
	cursor.All(context.TODO(), &users)
	if len(users) > 1 { // Encode ass array
		//Encode all users
		json.NewEncoder(w).Encode(users)
	} else { // Encode as single entry
		json.NewEncoder(w).Encode(users[0])
	}

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
		return
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
		return
	}
	var user User
	// decode json in request body
	err = json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	var doc bson.D
	data, _ := bson.Marshal(user)
	_ = bson.Unmarshal(data, &doc)
	// update specified user
	result, err := connection.Users.UpdateOne(
		context.TODO(),          // required context
		bson.M{"_id": objectId}, // filter
		bson.M{"$set": doc},
	)
	if err != nil {
		log.Println("Update Failed")
		return
	}
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
