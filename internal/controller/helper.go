package controller

import (

	"net/http"
	"PrismX/internal/database"
	"PrismX/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
)


// helper to build a filter for _id that accepts hex ObjectIDs or raw string ids
func buildIDFilter(id string) bson.M {
	if id == "" {
		return bson.M{}
	}
	if objID, err := bson.ObjectIDFromHex(id); err == nil {
		return bson.M{"_id": objID}
	}
	return bson.M{"_id": id}
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
