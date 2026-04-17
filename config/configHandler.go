package config

import (
    "sync"
)

// attributes are private to struct
// configs has to be added later via db
type upstream struct {
	servers  map[string]upstreamservers
}

type config struct {
	upstream upstream
	balancer  string
	server server
}

type upstreamservers struct{
	Address []string
	lbmethod string
	weight int
	replicas  int
	maxFails int
	FailTimeout int
	down bool
}

type server struct{
	serverName string
	count int
}

var cfg *config
var once sync.Once

// Dummy configuration for now
func LoadConfig() (*config, error) {

	once.Do(func() {
		cfg = &config{		
			upstream: upstream{
				servers: map[string]upstreamservers{
					"auth": {
						Address: []string{
							"http://localhost:9000/auth",
							"http://localhost:9001/auth",
						},
						lbmethod: "consistent-hash",
						weight:      10,
						maxFails:    2,
						replicas:    3,
						FailTimeout: 2,
						down:        false,
					},
					"order": {
						Address: []string{
							"http://localhost:9002/auth",
							"http://localhost:9003/auth",
						},
						lbmethod: "consistent-hash",
						weight:      10,
						maxFails:    2,
						replicas:    3,
						FailTimeout: 2,
						down:        false,
					},
				},
			},
			server: server{
				serverName: "RetailService",
				count:2,
			},
		}
	})

	return cfg, nil
}

// methods are public to export 
func (c *config) GetServers() map[string]upstreamservers {
	return c.upstream.servers
}

func(c *config) GetServerCount() int{
	return c.server.count
}

func(c * config) GetLbtype(service_name string) string{
	return c.upstream.servers[service_name].lbmethod
}
