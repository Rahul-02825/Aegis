package proxy

import (
	"PrismX/config"
	"PrismX/loadBalancer"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"fmt"
	// "PrismX/logger"
)

func StartProxy(cfg *config.Configs) {

	// cfg, _ := config.LoadConfig()
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
		path := r.URL.Path

		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 {
			http.Error(w, "Invalid route", http.StatusBadRequest)
			return
		}

		service := parts[0]
		fmt.Println(service)
		//  Check if service exists in config
		servers := cfg.GetServers()
		if _, ok := servers[service]; !ok {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		// Get load balancer
		balancers := loadBalancer.GetBalancers()
		if balancers == nil {
			http.Error(w, "Load balancers not initialized", http.StatusInternalServerError)
			return
		}

		lb, ok := balancers[service]
		if !ok {
			http.Error(w, "Load balancer missing", http.StatusInternalServerError)
			return
		}

		//  Key for hashing (or RR)
		key := r.RemoteAddr
		fmt.Println(key)
		//  Get target server dynamically
		target := lb.GetServer(key)
		fmt.Println(target)
		targetURL, err := url.Parse(target)
		if err != nil {
			http.Error(w, "Invalid upstream", http.StatusInternalServerError)
			return
		}

		// Rewrite path (remove service prefix)
		// r.URL.Path = "/" + strings.Join(parts[1:], "/")
		// Proxy
		fmt.Println(targetURL.String() + r.URL.Path)
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ServeHTTP(w, r)
	})

	http.ListenAndServe(":8080", nil)
}