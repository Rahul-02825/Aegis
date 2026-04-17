package main

import (
	// "PrismX/internal/controller"
	"PrismX/config"
	"PrismX/internal/controller"
	"PrismX/internal/database"
	"PrismX/loadBalancer"
	"PrismX/logger"
	"PrismX/proxy"
	// "fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	// proxy_mux := mux.NewRouter() 
	internal_mux := mux.NewRouter()
	log := logger.InitLogger("app.log")

	defer log.Close()

	log.Info("Application started")

	log.Info("Running database connection")
	database.ConnectDatabase()

	log.Info("Initialize configurations")

	cfg,err:=config.LoadConfig("69e275e4bf8504b6b004cf0b")
	if err != nil {
		log.Error("Error loading config: " + err.Error())
		return
	}
	log.Info("Config loaded successfully")
	
	loadBalancer.InitLoadBalancer(cfg)
	logger.Instance.Info("Initialized load balancers")

	// ------------------------------ Setting up HTTP routes -----------------------------

	log.Info("Setting up HTTP routes")

	// ----------------------------- Internal server routes ------------------------------
	// config routes
	internal_mux.HandleFunc("/createConfig",controller.CreateConfig)
	internal_mux.HandleFunc("/getConfigs",controller.GetConfigs)
	internal_mux.HandleFunc("/updateConfig",controller.UpdateConfig)

	// user routes
	internal_mux.HandleFunc("/createuser",controller.CreateUser)
	internal_mux.HandleFunc("/getusers",controller.GetUsers)
	internal_mux.HandleFunc("/getuser",controller.GetUser)
	internal_mux.HandleFunc("/updateuser",controller.UpdateUser)
	
	
	// proxy_server := &http.Server{
	// 	Addr:":8080",
	// 	Handler: proxy_mux,
	// }
	internal_server := &http.Server{
		Addr : ":8081",
		Handler : internal_mux,
	}
	
	// // cg ,err := config.DB_config()
	// if err != nil {
	// 	log.Error("Error loading config: " + err.Error())
	// 	return
	// }
	log.Info("Loaded config from DB successfully")
	// fmt.Printf("Config: %+v\n", cg)
	go internal_server.ListenAndServe()	
	// proxy_server.ListenAndServe()
	log.Info("Proxy server is running on port :8080")
	proxy.StartProxy(cfg)
}
