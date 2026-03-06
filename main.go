package main

import (
	"PrismX/internal/controller"
	"PrismX/internal/database"
	"PrismX/loadBalancer"
	"PrismX/logger"
	"net/http"
	"time"
	"github.com/gorilla/mux"
)

func main() {

	proxy_mux := mux.NewRouter() 
	internal_mux := mux.NewRouter()
	log := logger.InitLogger("app.log")

	defer log.Close()

	log.Info("Application started")

	log.Info("Running database connection")
	database.ConnectDatabase()

	go loadBalancer.StartLoadBalancer()

	log.Info("Setting up HTTP routes")
	proxy_mux.HandleFunc("/createuser",controller.CreateUser)
	proxy_mux.HandleFunc("/getusers",controller.GetUsers)
	proxy_mux.HandleFunc("/getuser",controller.GetUser)
	proxy_mux.HandleFunc("/updateuser",controller.UpdateUser)
	// http.HandleFunc("/deleteuser",controller.DeleteUser)
	internal_mux.HandleFunc("/start",func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("Internal server is running"))
	})
	
	for i := 0; i < 5; i++ {
		log.Info("Main thread working...")
		time.Sleep(1 * time.Second)
	}
	proxy_server := &http.Server{
		Addr:":8080",
		Handler: proxy_mux,
	}
	internal_server := &http.Server{
		Addr : ":8081",
		Handler : internal_mux,
	}
	log.Info("Proxy server is running on port :8080")
	
	go internal_server.ListenAndServe()
	proxy_server.ListenAndServe()
	

}
