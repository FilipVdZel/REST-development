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

// User type struct
type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Surname  string             `json:"surname,omitempty" bson:"surname,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Username string             `json:"username,omitempty" bson:"username,omitempty"`
	Password string             `json:"-" bson:"-"`
	Dob      string             `json:"dob,omitempty" bson:"dob,omitempty"`
}

// shorter version of user stored in Subsciptions collection
type ShortUser struct {
	Username string `json:"username,omitempty" bson:"username,omitempty"`
	Email    string `json:"email,omitempty" bson:"email,omitempty"`
}

// Messages saved on Channel
type Message struct {
	Message     string
	TimeCreated string
}

// Subscriptions type struct
type Subscription struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Owner       string             `json:"owner,omitempty" bson:"owner,omitempty"`
	OwnerEmail  string             `json:"owneremail,omitempty" bson:"owneremail,omitempty"`
	Subscribers []ShortUser        `json:"subscribers,omitempty" bson:"subscribers,omitempty"`
	Messages    []Message          `json:"messages,omitempty" bson:"messages,omitempty"`
}

// Database connection struct
type Connection struct {
	Subscriptions *mongo.Collection
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

	collectionSubscriptions := client.Database("myDB").Collection("Subscriptions")
	connection := Connection{
		Subscriptions: collectionSubscriptions,
	}

	// init server mux
	router := mux.NewRouter()

	//Handelers
	router.HandleFunc("/subscriptions", connection.getSubscriptions).Methods("GET")
	router.HandleFunc("/subscriptions", connection.createSubscriptions).Methods("POST")
	router.HandleFunc("/subscriptions/{id}", connection.updateSubscriptions).Methods("PUT")
	router.HandleFunc("/subscriptions/{id}", connection.deleteSubscriptions).Methods("DELETE")
	router.HandleFunc("/messages", connection.sendMessages).Methods("POST")
	router.HandleFunc("/subscribe/{id}", connection.Subscribe).Methods("POST")
	router.HandleFunc("/unsubscribe/{id}", connection.Unsubscribe).Methods("DELETE")

	// listen and serve requests on localhost port 8082
	// Use server mux router
	log.Fatal(http.ListenAndServe(":8082", router))
}

//Handlers
func (connection Connection) getSubscriptions(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")
	var subscriptions []Subscription

	//Get parameters value
	params := req.URL.Query()
	// If no parameters Encode all users

	if len(params) == 0 {
		// create a filter (cursor)
		// this filter returns all entries in the database
		cursor, err := connection.Subscriptions.Find(context.TODO(), bson.M{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `" }`))
			return
		}
		//Apply filter to all entries
		cursor.All(context.TODO(), &subscriptions)

		// repond with filtered content
		json.NewEncoder(w).Encode(subscriptions)
		return
	}

	//Go through all names
	searchName := req.URL.Query().Get("name")
	filter := bson.D{primitive.E{Key: "name", Value: primitive.Regex{Pattern: searchName, Options: "i"}}}
	cursor, err := connection.Subscriptions.Find(context.TODO(), filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `" }`))
		return
	}

	//Apply filter to all entries
	cursor.All(context.TODO(), &subscriptions)
	//Encode all Subscriptions
	json.NewEncoder(w).Encode(subscriptions)

}

func (connection Connection) createSubscriptions(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Create new user var and decode json contect from body
	var channel Subscription
	_ = json.NewDecoder(req.Body).Decode(&channel)

	// Ger username and password
	u, p, ok := req.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}
	// Confirm that user and password is correct
	valid := verifyUserPassword(u, p)
	if !valid {
		fmt.Println("Username and password not correct")
		return
	}
	channel.Owner = u
	var user User
	getUserDetails(u, &user)
	channel.OwnerEmail = user.Email
	// insert channel into database
	result, _ := connection.Subscriptions.InsertOne(context.TODO(), channel)
	//Response with json data testing
	json.NewEncoder(w).Encode(result)
	// Confirm that channel was created
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Channel " + channel.Name + " was created by user " + channel.Owner + "\n"))
	return

}

func (connection Connection) updateSubscriptions(w http.ResponseWriter, req *http.Request) {
	// make sure content is not served as text to client
	w.Header().Set("Content-Type", "application/json")
	// Confirm that user and password is correct
	u, p, ok := req.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}
	valid := verifyUserPassword(u, p)
	if !valid {
		fmt.Println("Username and password not correct")
		return
	}

	// retrieve map of veriables from get url
	param := mux.Vars(req)
	// Gat object ID
	objectId, err := primitive.ObjectIDFromHex(param["id"])
	if err != nil {
		log.Println("Invalid ID")
		return
	}
	var channel Subscription
	// decode json in request body
	err = json.NewDecoder(req.Body).Decode(&channel)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	// Check if user is the owner of the Channel
	var channeldata Subscription
	connection.Subscriptions.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&channeldata)
	if u != channeldata.Owner {
		fmt.Println("Not the owner")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Permission Denied.\n"))
		return
	}

	// Update Channel info
	var doc bson.D
	data, _ := bson.Marshal(channel)
	_ = bson.Unmarshal(data, &doc)
	// update specified user
	result, err := connection.Subscriptions.UpdateOne(
		context.TODO(),          // required context
		bson.M{"_id": objectId}, // filter
		bson.M{"$set": doc},
	)
	if err != nil {
		log.Println("Update Failed")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Update failed\n"))
		return
	}
	json.NewEncoder(w).Encode(result)

}

func (connection Connection) deleteSubscriptions(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Confirm that user and password is correct
	u, p, ok := req.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}
	valid := verifyUserPassword(u, p)
	if !valid {
		fmt.Println("Username and password not correct")
		return
	}

	//Create new user var and decode json contect from body
	param := mux.Vars(req)
	//Get Object id
	objectId, err := primitive.ObjectIDFromHex(param["id"])
	if err != nil {
		log.Println("Invalid ID")
	}
	// Check if user is the owner of the Channel
	var channel Subscription
	connection.Subscriptions.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&channel)
	if u != channel.Owner {
		fmt.Println("Not the owner")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Permission Denied.\n"))
		return
	}
	// Delete Channel from collection
	result, err := connection.Subscriptions.DeleteOne(context.TODO(), bson.M{"_id": objectId})

	//Response with json data
	json.NewEncoder(w).Encode(result)

}

func (connection Connection) sendMessages(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	// Confirm that user and password is correct
	u, p, ok := req.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}
	valid := verifyUserPassword(u, p)
	if !valid {
		fmt.Println("Username and password not correct")
		return
	}

	//Get parameters value
	params := req.URL.Query()
	// Get channel name
	searchChannel := params.Get("channel")

	// Check if user is the owner of the Channel
	var channel Subscription
	connection.Subscriptions.FindOne(context.TODO(), bson.M{"name": searchChannel}).Decode(&channel)
	fmt.Println("Username: " + u + ", Owner: " + channel.Owner)
	if u != channel.Owner {
		fmt.Println("Not the owner")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Permission Denied.\n"))
		return
	}

	// TODO: Change encoding to HTML
	var message Message
	err := json.NewDecoder(req.Body).Decode(&message)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	// Add time to message
	currentTime := time.Now()
	t := currentTime.Format("2006-01-02 15:04:05")
	message.TimeCreated = t

	// Insert message as embedded document
	var doc bson.D
	data, _ := bson.Marshal(message)
	_ = bson.Unmarshal(data, &doc)
	result, err := connection.Subscriptions.UpdateOne(
		context.TODO(),
		bson.M{"_id": channel.ID},
		bson.M{"$push": bson.M{"Messages": doc}},
	)
	if err != nil {
		log.Printf("%v\n", err)
	}
	log.Printf("%+v\n", result)

	// Send message to all subscribers
	ownerEmail := channel.OwnerEmail
	messageText := message.Message
	w.Header().Set("Content-Type", "text/plain")
	for _, subs := range channel.Subscribers {
		// Send actual emails here
		userEmail := subs.Email
		username := subs.Username
		w.Write([]byte(ownerEmail + " sent message to " + username + " at " +
			userEmail + " message: " + messageText + "\n"))
	}

}

func (connection Connection) Subscribe(w http.ResponseWriter, req *http.Request) {
	//Get parameters value
	params := req.URL.Query()
	// If no parameters Give error
	if len(params) == 0 {
		log.Panicln("No Username Given")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("No Username Given. Add ?username=username to url"))
		return
	}
	// Get user details from User server
	username := params.Get("username")
	var user User
	getUserDetails(username, &user)
	if user.Username == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Username " + username + " not found\n"))
		return
	}
	shortuser := ShortUser{
		Username: user.Username,
		Email:    user.Email,
	}
	// Get channel document from collection
	param := mux.Vars(req)
	objectId, err := primitive.ObjectIDFromHex(param["id"])
	if err != nil {
		log.Println("Invalid ID")
	}
	// Get Channel
	var channel Subscription
	connection.Subscriptions.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&channel)

	// Insert shortUser as embedded document
	var doc bson.D
	data, _ := bson.Marshal(shortuser)
	_ = bson.Unmarshal(data, &doc)
	connection.Subscriptions.UpdateOne(
		context.TODO(),
		bson.M{"_id": objectId},
		bson.M{"$push": bson.M{"Subscribers": doc}},
	)

	// Send back response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User " + username + " successfully subscribed to " + channel.Name + "\n"))

}

func (connection Connection) Unsubscribe(w http.ResponseWriter, req *http.Request) {
	//Get parameters value
	params := req.URL.Query()
	// If no parameters Give error
	if len(params) == 0 {
		log.Panicln("No Username Given")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("No Username Given. Add ?username=username to url"))
		return
	}
	username := params.Get("username")

	// Get channel document from collection
	param := mux.Vars(req)
	objectId, err := primitive.ObjectIDFromHex(param["id"])
	if err != nil {
		log.Println("Invalid ID")
	}

	connection.Subscriptions.UpdateOne(
		context.TODO(),
		bson.M{"_id": objectId},
		bson.M{"$pull": bson.M{"Subscribers": bson.M{"username": username}}},
	)

	// Send back response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User " + username + " successfully unsubscribed\n"))

}

func verifyUserPassword(username string, password string) bool {
	url := "http://server-users:8081/verifyUser"
	method := "POST"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("Got error %s\n", err.Error())
		return false
	}
	req.SetBasicAuth(username, password)
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
		return false
	}
	defer response.Body.Close()
	status := response.StatusCode
	if status == 200 {
		return true
	} else {
		return false
	}

}

func getUserDetails(username string, user *User) {
	// Set up request
	url := "http://server-users:8081/users?username=" + username
	method := "GET"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("Got error %s\n", err.Error())
		return
	}
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
		return
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(user)
	if err != nil {
		log.Printf("%v\n", err)
	}
	return

}
