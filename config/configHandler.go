package config


// attributes are private to struct
// configs has to be added later via db
type upstream struct {
	name     string
	lbMethod string
	servers  map[string]upstreamservers
	replicas int
}

type config struct {
	upstream upstream
	balancer  string
	server []server
}
type upstreamservers struct{
	Address string
	weight int
	replicas  int
	maxFails int
	FailTimeout int
	down bool
}

type server struct{
	serverName string
}

// Dummy configuration for now
func LoadConfig() (*config, error) {
	return &config{
		upstream: upstream{
				name:     "Auth service",
				lbMethod: "consistent-hash",
				
				servers:  map[string]upstreamservers{
					"auth":{
						Address:"http://localhost:9000/auth",
						weight:10,
						maxFails: 2,
						replicas: 3,
						FailTimeout: 2,
						down:false,
					},
					"order":{
						Address:"http://localhost:9001/order",
						weight:10,
						maxFails: 2,
						replicas: 3,
						FailTimeout: 2,
						down:false,
					},
				},
			},
		server:[]server{
			{serverName:"RetailService"},
		},
	}, nil
}

// methods are public to export 
func (c *config) GetServers() map[string]upstreamservers {
	return c.upstream.servers
}
