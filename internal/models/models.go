
package models
import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Password string             `bson:"password" json:"password"`
}

type Config struct {
	ID        primitive.ObjectID       `bson:"_id,omitempty" json:"id"`
	Upstreams map[string]Upstream      `bson:"upstreams" json:"upstreams"`
	Servers   []Server                 `bson:"servers" json:"servers"`
	Global    GlobalConfig             `bson:"global" json:"global"`
}

type Upstream struct {
	Name     string             `bson:"name" json:"name"`
	LBMethod string             `bson:"lb_method" json:"lb_method"`
	Servers  []UpstreamServer   `bson:"servers" json:"servers"`
	Zone     string             `bson:"zone" json:"zone"`
}

type UpstreamServer struct {
	Address     string             `bson:"address" json:"address"`
	Weight      int                `bson:"weight,omitempty" json:"weight,omitempty"`
	MaxFails    int                `bson:"max_fails,omitempty" json:"max_fails,omitempty"`
	FailTimeout string             `bson:"fail_timeout,omitempty" json:"fail_timeout,omitempty"`
	Backup      bool               `bson:"backup,omitempty" json:"backup,omitempty"`
	Down        bool               `bson:"down,omitempty" json:"down,omitempty"`
}

type Server struct {
	ServerName string               `bson:"server_name" json:"server_name"`
	Listen     int                  `bson:"listen" json:"listen"`
	SSL        bool                 `bson:"ssl" json:"ssl"`
	CertPath   string               `bson:"cert_path,omitempty" json:"cert_path,omitempty"`
	KeyPath    string               `bson:"key_path,omitempty" json:"key_path,omitempty"`
	Locations  []Location           `bson:"locations" json:"locations"`
	Headers    map[string]string    `bson:"headers,omitempty" json:"headers,omitempty"`
}

type Location struct {
	Path         string             `bson:"path" json:"path"`
	ProxyPass    string             `bson:"proxy_pass" json:"proxy_pass"`
	Websocket    bool               `bson:"websocket" json:"websocket"`
	Headers      map[string]string  `bson:"headers,omitempty" json:"headers,omitempty"`
	RedirectHTTP bool               `bson:"redirect_http_to_https,omitempty" json:"redirect_http_to_https,omitempty"`
}

type GlobalConfig struct {
	KeepaliveTimeout   string             `bson:"keepalive_timeout" json:"keepalive_timeout"`
	ClientMaxBodySize  string             `bson:"client_max_body_size" json:"client_max_body_size"`
	AccessLog          string             `bson:"access_log" json:"access_log"`
	ErrorLog           string             `bson:"error_log" json:"error_log"`
}