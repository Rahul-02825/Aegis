package controller

import (
	"PrismX/internal/database"
	"PrismX/internal/models"
	"PrismX/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CreateConfig inserts a new Config document
func CreateConfig(res http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method)
	if req.Method != http.MethodPost {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res.Header().Set("Content-Type", "application/json")

	var cfg models.Config
	if err := json.NewDecoder(req.Body).Decode(&cfg); err != nil {
		logger.Instance.Error("Error decoding json for config")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if !ensureConfigCollection(res) {
		return
	}
	result, err := database.ConfigCollection.InsertOne(req.Context(), cfg)
	if err != nil {
		logger.Instance.Error("Server error in creating config: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Instance.Info("New config created successfully")
	json.NewEncoder(res).Encode(result)
}

// GetConfigs returns all configs or single config if `id` query param provided
func GetConfigs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	id := req.URL.Query().Get("id")
	if id != "" {
		var result bson.M
		filter := buildIDFilter(id)
		if !ensureConfigCollection(res) {
			return
		}
		if err := database.ConfigCollection.FindOne(context.Background(), filter).Decode(&result); err != nil {
			logger.Instance.Error("DB error finding config: " + err.Error())
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(res).Encode(result)
		return
	}
	if !ensureConfigCollection(res) {
		return
	}
	cursor, err := database.ConfigCollection.Find(context.Background(), bson.M{})
	if err != nil {
		logger.Instance.Error("DB error finding configs: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		logger.Instance.Error("DB cursor error for configs: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(results)
}

// UpdateConfig updates an existing config. Provide `id` query param or `id` in body.
func UpdateConfig(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	res.Header().Set("Content-Type", "application/json")

	id := req.URL.Query().Get("id")
	var body map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Instance.Error("Error decoding request body for UpdateConfig: " + err.Error())
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
	delete(body, "id")
	delete(body, "_id")
	if !ensureConfigCollection(res) {
		return
	}
	filter := buildIDFilter(id)
	update := bson.M{"$set": body}
	result, err := database.ConfigCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.Instance.Error("DB error updating config: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(result)
}

