package config

import (
	"PrismX/internal/database"
	"PrismX/internal/models"
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// attributes are private to struct
// configs has to be added later via db
type upstream struct {
	servers  map[string]upstreamservers
}

type Configs struct {
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


var (
	cfg *Configs
	mu  sync.RWMutex
)
// Dummy configuration for now
// func LoadConfig() (*config, error) {

// 	once.Do(func() {
// 		cfg = &config{		
// 			upstream: upstream{
// 				servers: map[string]upstreamservers{
// 					"auth": {
// 						Address: []string{
// 							"http://localhost:9000",
// 							"http://localhost:9001",
// 						},
// 						lbmethod: "consistent-hash",
// 						weight:      10,
// 						maxFails:    2,
// 						replicas:    3,
// 						FailTimeout: 2,
// 						down:        false,
// 					},
// 					"order": {
// 						Address: []string{
// 							"http://localhost:9002",
// 							"http://localhost:9003",
// 						},
// 						lbmethod: "consistent-hash",
// 						weight:      10,
// 						maxFails:    2,
// 						replicas:    3,
// 						FailTimeout: 2,
// 						down:        false,
// 					},
// 				},
// 			},
// 			server: server{
// 				serverName: "RetailService",
// 				count:2,
// 			},
// 		}
// 	})

// 	return cfg, nil
// }

func LoadConfig(id string) (*Configs, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dbConfig models.Config

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = database.ConfigCollection.
		FindOne(ctx, bson.M{"_id": objID}).
		Decode(&dbConfig)

	if err != nil {
		return nil, err
	}

	newCfg := convertToRuntimeConfig(dbConfig)

	mu.Lock()
	cfg = newCfg
	mu.Unlock()

	return newCfg, nil
}

func convertToRuntimeConfig(db models.Config) *Configs {

	upstreamMap := make(map[string]upstreamservers)

	for name, up := range db.Upstreams {

		var addresses []string

		for _, s := range up.Servers {
			if s.Down {
				continue
			}
			addresses = append(addresses, s.Address)
		}

		upstreamMap[name] = upstreamservers{
			Address:     addresses,
			lbmethod:    up.LBMethod,
			weight:      up.Servers[0].Weight,
			maxFails:    up.Servers[0].MaxFails,
			replicas:    3, // or derive later
			FailTimeout: 2, // convert from string later
			down:        false,
		}
	}

	return &Configs{
		upstream: upstream{
			servers: upstreamMap,
		},
		server: server{
			serverName: db.Servers[0].ServerName,
			count:      len(db.Servers),
		},
	}
}


// methods are public to export 
func (c *Configs) GetServers() map[string]upstreamservers {
	return c.upstream.servers
}

func(c *Configs) GetServerCount() int{
	return c.server.count
}

func(c * Configs) GetLbtype(service_name string) string{
	return c.upstream.servers[service_name].lbmethod
}

func GetConfig() *Configs {
	mu.RLock()
	defer mu.RUnlock()
	return cfg
}