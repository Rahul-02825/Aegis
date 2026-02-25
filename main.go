package main

import (
	"time"

	"PrismX/logger"
	"PrismX/loadBalancer"
	"PrismX/internal/database"
	"PrismX/internal/controller"
	"net/http"

)

func main() {

	log := logger.InitLogger("app.log")

	defer log.Close()

	log.Info("\nApplication started")

	log.Info("Running database connection")
	database.ConnectDatabase()

	go loadBalancer.StartLoadBalancer()

	http.HandleFunc("/createuser",controller.CreateUser)
	http.HandleFunc("/getusers",controller.GetUsers)
	http.HandleFunc("/getuser",controller.GetUser)
	http.HandleFunc("/updateuser",controller.UpdateUser)
	// http.HandleFunc("/deleteuser",controller.DeleteUser)
	
	for i := 0; i < 5; i++ {
		log.Info("Main thread working...")
		time.Sleep(1 * time.Second)
	}

	log.Info("Proxy server is running on port :8080")
	http.ListenAndServe(":8080", nil)
}
