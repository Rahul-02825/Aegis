package loadBalancer

import (
	"fmt"
	"PrismX/config"
	"PrismX/logger"

)

func StartLoadBalancer() {
	logger.Instance.Info("Starting load balancer")

	// Factory decides which algorithm to use
	lb,err:= balancerFactory("consistent-hash")
	// if err != nil {
	// 	log.Instance.Error(err.Error())
	// 	return
	// }

	// Load servers from config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Instance.Error("failed to load config")
		return
	}

	servers := cfg.GetServers()
	fmt.Println(servers["auth"])
	for _, server := range servers {
		fmt.Println(server.Address)
		lb.insertServer(server.Address)
	}

	// Example requests 
	reqs := []string{"request1", "request2", "request3"}

	for _, r := range reqs {
		server := lb.getServer(r)
		fmt.Printf("Request %s → %s\n", r, server)
	}

	// Dynamic removal example
	lb.removeServer("server2")
	logger.Instance.Warn("server2 removed")

	fmt.Println("After removal:")
	fmt.Println(lb.getServer("request1"))
}
