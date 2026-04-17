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

func InitLoadBalancer(cfg *config.Configs) {

	once.Do(func() {

		logger.Instance.Info("Initializing load balancers")

		// cfg, err := config.LoadConfig()
		// if err != nil {
		// 	logger.Instance.Error("Failed to load config")
		// 	return
		// }

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



func GetBalancer(service_name string) (Loadbalancer, bool) {

	// if balancers == nil {
	// 	InitLoadBalancer()
	// }

	lb, ok := balancers[service_name]
	return lb, ok
}

