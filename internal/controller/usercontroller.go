package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"PrismX/internal/database"
	"PrismX/internal/models"
	"PrismX/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
)
	
func CreateUser(res http.ResponseWriter,req *http.Request){
	
	// Make sure it is a POST request
	if req.Method != http.MethodPost {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	res.Header().Set("Content-Type", "application/json")

	var user models.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil{
		logger.Instance.Error("Error in decoding json from request(user controller)")
		http.Error(res,err.Error(),http.StatusBadRequest)
		return 
	}
	result,err := database.UserCollection.InsertOne(req.Context(),user)
	
	if err != nil{
		logger.Instance.Error("Server error in creating user")
		http.Error(res,err.Error(),500)
		return
	}
	logger.Instance.Info("New user created successfully")
	json.NewEncoder(res).Encode(result)

}


// GetUsers returns all users
func GetUsers(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res.Header().Set("Content-Type", "application/json")

	if !ensureUserCollection(res) {
		return
	}

	cursor, err := database.UserCollection.Find(req.Context(), bson.M{})
	if err != nil {
		logger.Instance.Error("DB error finding users: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		logger.Instance.Error("DB cursor error for users: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(results)
}

// GetUser returns a single user by query param `id`
func GetUser(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	id := req.URL.Query().Get("id")
	if id == "" {
		http.Error(res, "missing id query param", http.StatusBadRequest)
		return
	}
	var result bson.M
	filter := buildIDFilter(id)
	if !ensureUserCollection(res) {
		return
	}
	if err := database.UserCollection.FindOne(context.Background(), filter).Decode(&result); err != nil {
		logger.Instance.Error("DB error finding user: " + err.Error())
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(res).Encode(result)
}

// UpdateUser updates fields of a user. Provide JSON body; `id` must be present in body or query param.
func UpdateUser(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res.Header().Set("Content-Type", "application/json")

	// try to get id from query or body
	id := req.URL.Query().Get("id")
	var body map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Instance.Error("Error decoding request body for UpdateUser: " + err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if id == "" {
		if v, ok := body["id"]; ok {
			if s, ok2 := v.(string); ok2 {
				id = s
			}
		}
		if id == "" {
			http.Error(res, "missing id for update", http.StatusBadRequest)
			return
		}
	}
	// remove immutable fields
	delete(body, "id")
	delete(body, "_id")

	if !ensureUserCollection(res) {
		return
	}
	filter := buildIDFilter(id)
	update := bson.M{"$set": body}
	result, err := database.UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.Instance.Error("DB error updating user: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(result)
}

