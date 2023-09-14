package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// struct for storing data
type User struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty"`
	Email string             `json:"email,omitempty" bson:"email,omitempty"`
	City  string             `json:"city,omitempty" bson:"city,omitempty"`
}

var userCollection = db().Database("myproject").Collection("user") // get collection "users" from db() which returns *mongo.Client

// Create Profile or Signup

func createProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json") // for adding Content-type

	// Declare a User struct variable named person.
	var person User

	// Use the json.NewDecoder() function to decode the JSON request body from the HTTP request
	//and store it in the person variable.
	err := json.NewDecoder(r.Body).Decode(&person) // storing in person variable of type user
	if err != nil {
		fmt.Print(err)
	}

	//Insert the person variable into the MongoDB collection using the InsertOne() method
	result, err := userCollection.InsertOne(context.TODO(), person)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", result)

	// Encode the result.InsertedID into a JSON format using the json.NewEncoder() function and
	// write it to the response using the Encode() method.
	//This returns the MongoDB ID of the newly created document to the client.
	json.NewEncoder(w).Encode(result.InsertedID) // return the mongodb ID of generated document

}

// Get Profile of a particular User by Name

func getUserProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	// get the id from request r
	params := mux.Vars(r)
	id := params["_id"]

	person := User{}

	// convert the user id from hexadecimal string to mongoDB objcectID
	objID, errObjID := primitive.ObjectIDFromHex(id)
	if errObjID != nil {
		fmt.Println(errObjID)
	}
	//Create a filter that matches the user ID using the bson.M{} function
	filter := bson.M{"_id": objID}

	//Use the FindOne() method of the MongoDB collection to retrieve the user document
	// that matches the filter.
	err := userCollection.FindOne(context.TODO(), filter).Decode(&person)
	if err != nil {
		fmt.Println(err)
	}
	// Encode the User struct into a JSON format using the json.NewEncoder() function
	//and write it to the response using the Encode() method.
	json.NewEncoder(w).Encode(person) // returns a Map containing document

}

// update the users name.

func updateProfileName(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request URL
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["_id"])
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	// Decode the request body into a User struct
	var person User
	err = json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		fmt.Print(err)
	}
	filter := bson.D{{"_id", id}} // converting value to BSON type
	after := options.After        // for returning updated document
	returnOpt := options.FindOneAndUpdateOptions{

		ReturnDocument: &after,
	}
	update := bson.D{{"$set", bson.D{{"name", person.Name}}}}
	updateResult := userCollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)

	var result User
	_ = updateResult.Decode(&result)

	json.NewEncoder(w).Encode(result)
}

// Update teh user profile

func aaupdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request URL
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["_id"])
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	// Decode the request body into a User struct
	var person User
	err = json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		fmt.Print(err)

	}
	filter := bson.D{{"_id", id}} // converting value to BSON type
	options := options.FindOneAndReplace().SetReturnDocument(options.After)
	update := bson.M{"$set": bson.M{
		"name":  person.Name,
		"email": person.Email,
		"city":  person.City,
	}}

	updateResult := userCollection.FindOneAndReplace(context.TODO(), filter, update, options)

	var result User
	_ = updateResult.Decode(&result)

	fmt.Println("updated document:", result)

	json.NewEncoder(w).Encode(result)
}

//delete the user profile

func deleteProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["_id"])
	filter := bson.M{"_id": id}
	deletedresult, err := userCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the user was deleted successfully
	if deletedresult.DeletedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(deletedresult) // return number of documents deleted
}

// get all users

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var retrievedusers []User                                   //slice for multiple documents
	cur, err := userCollection.Find(context.TODO(), bson.D{{}}) //returns a *mongo.Cursorif
	if err != nil {
		log.Fatal(err)
		fmt.Printf(err.Error())
	}

	for cur.Next(context.TODO()) { //Next() gets the next document for corresponding cursor
		var user User
		err := cur.Decode(&user)
		if err != nil {
			log.Fatal(err)
		}

		retrievedusers = append(retrievedusers, user) // appending document pointed by Next()

	}
	if err := cur.Close(context.TODO()); err != nil {
		log.Fatal(err)
	}
	// close the cursor once stream of documents has exhausted

	json.NewEncoder(w).Encode(retrievedusers)

}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON payload from the request body
	var person User
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert the ID from a string to a primitive.ObjectID
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the filter to find the user by their ID
	filter := bson.M{"_id": id}

	// Create the replacement document with the updated user information
	replace := bson.M{
		"name":  person.Name,
		"email": person.Email,
		"city":  person.City,
	}

	var user User
	err = userCollection.FindOneAndReplace(context.TODO(), filter, replace).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the updated user as JSON and send it back in the response
	json.NewEncoder(w).Encode(user)
}
