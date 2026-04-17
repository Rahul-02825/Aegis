package loadBalancer

import (
	"PrismX/config"
	"PrismX/logger"
	"sync"

)

var (
	balancers map[string]Loadbalancer
	once sync.Once
)

func InitLoadBalancer() {

	once.Do(func() {

		logger.Instance.Info("Initializing load balancers")

		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Instance.Error("Failed to load config")
			return
		}

		balancers = make(map[string]Loadbalancer) 
		servers := cfg.GetServers()

		for service_name, upstream := range servers {

			service_lb := cfg.GetLbtype(service_name)
			lb, _ := balancerFactory(service_lb)

			for _, addr := range upstream.Address {
				lb.insertServer(addr)
			}

			balancers[service_name] = lb
		}
	})
}

// GetBalancers returns the initialized balancers map
func GetBalancers() map[string]Loadbalancer {
	return balancers
}

// StartLoadBalancer initializes the load balancers (for backward compatibility)
// func StartLoadBalancer() {
// 	InitLoadBalancer()
// }
// func InitLoadBalancer() {
// 	logger.Instance.Info("Starting load balancer")

// 	// Load servers from config
// 	logger.Instance.Info("Trying fetch Config:Start.go")
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		logger.Instance.Error("Failed to load config:Start.go")
// 		return
// 	}
// 	logger.Instance.Info("Config fetched Successfully:Start.go")

// 	// Factory decides which algorithm to use

// 	// balancers := make(map[string]Loadbalancer)
// 	servers := cfg.GetServers()

// 	for service_name, upstream := range servers {

// 		service_lb := cfg.GetLbtype(service_name)
// 		lb, _ := balancerFactory(service_lb)

// 		for _, addr := range upstream.Address {
// 			lb.insertServer(addr) // add servers into SAME ring
// 			fmt.Println(addr)
// 		}

// 		balancers[service_name] = lb
// 	}

	// if err != nil {
	// 	log.Instance.Error(err.Error())
	// 	return
	// }


	// // Example requests 
	// reqs := []string{"request1", "request2", "request3","fdifskdfjsdjf"}

	// for _, r := range reqs {
	// 	server := balancers["auth"].getServer(r)
	// 	fmt.Printf("Request %s → %s\n", r, server)
	// }

	// // Dynamic removal example
	// balancers["auth"].removeServer("server2")
	// logger.Instance.Warn("server2 removed")

	// fmt.Println("After removal:")
	// fmt.Println(balancers["auth"].getServer("request1"))
// }

func GetBalancer(service_name string) (Loadbalancer, bool) {

	if balancers == nil {
		InitLoadBalancer()
	}

	lb, ok := balancers[service_name]
	return lb, ok
}

