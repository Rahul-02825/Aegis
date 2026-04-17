package controller

import (

	"net/http"
	"PrismX/internal/database"
	"PrismX/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

)


// helper to build a filter for _id that accepts hex ObjectIDs or raw string ids
func buildIDFilter(id string) bson.M {
	if id == "" {
		return bson.M{}
	}

	objID, err := primitive.ObjectIDFromHex(id) // ✅ FIX
	if err != nil {
		return bson.M{}
	}

	return bson.M{"_id": objID}
}

// Ensure database collections are initialized to avoid nil-pointer panics
func ensureUserCollection(res http.ResponseWriter) bool {
	if database.UserCollection == nil {
		logger.Instance.Error("UserCollection is nil; database not initialized")
		http.Error(res, "database not initialized", http.StatusInternalServerError)
		return false
	}
	return true
}

func ensureConfigCollection(res http.ResponseWriter) bool {
	if database.ConfigCollection == nil {
		logger.Instance.Error("ConfigCollection is nil; database not initialized")
		http.Error(res, "database not initialized", http.StatusInternalServerError)
		return false
	}
	return true
}
