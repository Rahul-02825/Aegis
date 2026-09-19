# Aegis (module: PrismX)

A lightweight **reverse proxy / API gateway** written in Go, with dynamic upstream
configuration backed by MongoDB, pluggable load-balancing strategies, and a
built-in (not-yet-wired) rate limiter.

> **Status:** early development. Core proxy + config + load-balancer flow works
> end-to-end; several pieces (rate limiting, per-user auth, HTTPS, hot-reload
> across processes) are stubbed or partially wired. See [Known gaps](#known-gaps--wip) below.

---

## What it does

Aegis sits in front of a set of backend "upstream" services (e.g. `auth`,
`order`, `payment`). Incoming requests are routed by the **first path
segment** (`/auth/...`, `/order/...`) to the matching upstream group, and a
**consistent-hashing load balancer** picks which backend server in that group
handles the request. Upstream topology is not hardcoded — it's stored as a
**Config document in MongoDB** and loaded into memory at startup (and
reloaded whenever that config is updated via the internal API).

A separate internal HTTP API (port `8081`) lets you manage:
- **Configs** — upstream groups, servers, load-balancing method, per-server weight/health settings
- **Users** — basic user records (currently plaintext, see gaps)

---

## Architecture

```
                              ┌────────────────────────────┐
                              │          MongoDB            │
                              │  ┌────────────┐ ┌─────────┐ │
                              │  │   Config    │ │  User   │ │
                              │  └────────────┘ └─────────┘ │
                              └───────┬──────────────┬───────┘
                                      │              │
                     ┌────────────────┘              └───────────────┐
                     │                                                │
          ┌──────────▼───────────┐                        ┌──────────▼───────────┐
          │   config package     │                        │  internal/database    │
          │  LoadConfig(id)      │◄───────────────────────┤  Mongo client, User &  │
          │  in-memory *Configs  │                        │  Config collections    │
          └──────────┬───────────┘                        └────────────────────────┘
                     │ cfg
                     │
          ┌──────────▼───────────────┐
          │   loadBalancer package    │
          │  InitLoadBalancer(cfg)    │
          │  factory → ConsistentHash │
          │  map[service]Loadbalancer │
          └──────────┬────────────────┘
                     │ balancers
                     │
          ┌──────────▼───────────────────────────┐          ┌─────────────────────────────┐
Client ──►│         proxy package                  │         │   internal/controller       │
requests  │  StartProxy(cfg) — listens on :8080    │         │  (mux router on :8081)      │
          │  1. parse first path segment → service │         │  /createConfig  /getConfigs │
          │  2. look up upstream + load balancer    │         │  /updateConfig              │
          │  3. lb.GetServer(RemoteAddr) → target   │         │  /createuser /getusers      │
          │  4. httputil.ReverseProxy → target      │         │  /getuser /updateuser       │
          └────────────────────────────────────────┘         └──────────────┬──────────────┘
                                                                              │ on UpdateConfig,
                                                                              │ re-triggers
                                                                              │ config.LoadConfig(id)
                                                                              ▼
                                                                     (reloads in-memory cfg)

          ┌─────────────────────┐        ┌───────────────────────┐
          │   logger package    │        │  security package      │
          │  singleton file+     │        │  per-IP token bucket   │
          │  stdout logger       │        │  rate limiter           │
          │  ./logs/app.log      │        │  (standalone, NOT       │
          │                      │        │   wired into proxy yet) │
          └──────────────────────┘        └─────────────────────────┘
```

### Package responsibilities

| Package | File(s) | Responsibility |
|---|---|---|
| `main` | [main.go](main.go) | Wires everything together: init logger → connect DB → load config → init load balancers → start internal API (`:8081`, goroutine) → start proxy (`:8080`, blocking) |
| `logger` | [logger/logger.go](logger/logger.go) | Singleton (`sync.Once`) file + stdout logger, writes to `./logs/app.log` with `INFO`/`WARN`/`ERROR` levels |
| `internal/database` | [internal/database/db.go](internal/database/db.go) | Loads `.env` (`MONGO_URL`), connects to MongoDB, exposes `UserCollection` and `ConfigCollection` |
| `internal/models` | [internal/models/models.go](internal/models/models.go) | BSON/JSON schema for `User` and `Config` (upstreams, servers, locations, global settings) — this is the **on-disk / API shape** |
| `config` | [config/configHandler.go](config/configHandler.go) | Loads a `Config` doc by ID from Mongo, converts it into an unexported **runtime shape** (`Configs`), exposes read-only getters (`GetServers`, `GetLbtype`, `GetServerCount`); guarded by an `RWMutex` for concurrent reload |
| `loadBalancer` | [loadBalancer/*.go](loadBalancer) | `Loadbalancer` interface (`insertServer`, `removeServer`, `GetServer`) + factory (`balancerFactory`) that currently only builds `ConsistentHash`; `InitLoadBalancer` builds one balancer per upstream service from the runtime config |
| `proxy` | [proxy/reverse_proxy.go](proxy/reverse_proxy.go) | The actual reverse proxy handler: parses the service name out of the path, resolves it against config + load balancer, forwards via `httputil.NewSingleHostReverseProxy` |
| `internal/controller` | [internal/controller/*.go](internal/controller) | HTTP handlers for the internal management API — CRUD-ish endpoints for `Config` and `User`, backed directly by Mongo |
| `security` | [security/limiter.go](security/limiter.go) | Per-client-IP token-bucket rate limiter (`golang.org/x/time/rate`) with a background cleanup goroutine — implemented as its **own standalone `main()`**, not yet imported/used by `proxy` or the internal API |

---

## Data model

`Config` (stored in Mongo `PrismX.Config`, loaded via `config.LoadConfig(id)`):

```json
{
  "upstreams": {
    "auth": {
      "name": "auth",
      "lb_method": "consistent-hash",
      "zone": "auth_zone",
      "servers": [
        { "address": "http://localhost:9000", "weight": 10, "max_fails": 2, "fail_timeout": "2s", "down": false }
      ]
    }
  },
  "servers": [
    {
      "server_name": "RetailService",
      "listen": 8080,
      "ssl": false,
      "locations": [
        { "path": "/auth", "proxy_pass": "auth" }
      ]
    }
  ],
  "global": {
    "keepalive_timeout": "65s",
    "client_max_body_size": "10m",
    "access_log": "/var/log/access.log",
    "error_log": "/var/log/error.log"
  }
}
```

See [example_config.txt](example_config.txt) for full create/update payload examples.

`User` (stored in Mongo `PrismX.User`):

```json
{ "id": "...", "name": "...", "password": "..." }
```

---

## Workflow

### 1. Startup sequence ([main.go](main.go))

1. Init `mux.Router` for the internal API
2. Init logger → `./logs/app.log`
3. `database.ConnectDatabase()` — load `.env`, connect to Mongo, bind `UserCollection` / `ConfigCollection`
4. `config.LoadConfig("<hardcoded ObjectID>")` — fetch one `Config` doc, convert to runtime `Configs`
5. `loadBalancer.InitLoadBalancer(cfg)` — build one `ConsistentHash` balancer per upstream, seeded with its (non-`down`) server addresses
6. Register internal API routes and start `:8081` in a goroutine
7. `proxy.StartProxy(cfg)` — start the reverse proxy on `:8080` (blocking call, keeps the process alive)

### 2. Request routing (data plane, port `:8080`)

For every incoming request to the proxy:
1. Split the URL path; the **first segment is the service name** (e.g. `GET /auth/login` → service `auth`)
2. Look up that service in `cfg.GetServers()` — 404 if unknown
3. Look up the corresponding `Loadbalancer` from the in-memory `balancers` map
4. Call `lb.GetServer(r.RemoteAddr)` — the consistent-hash ring maps the client's remote address to a backend server address (same client tends to stick to the same backend)
5. Reverse-proxy the request to that backend via `httputil.NewSingleHostReverseProxy`

### 3. Config management (control plane, port `:8081`)

- `POST /createConfig` — insert a new `Config` document
- `GET /getConfigs[?id=...]` — list all configs, or fetch one by id
- `PUT /updateConfig?id=...` — `$set` partial update on a `Config` document, **then immediately calls `config.LoadConfig(id)` again** to hot-reload the in-memory runtime config used by the proxy

  Note: load balancers are **not** rebuilt on config reload — `loadBalancer.InitLoadBalancer` is guarded by `sync.Once` and only ever runs once at startup, so adding/removing upstream servers via `updateConfig` updates `config.GetConfig()` but does **not** currently propagate to the live hash rings used by the proxy.

- `POST /createuser`, `GET /getusers`, `GET /getuser?id=...`, `PUT /updateuser?id=...` — basic user CRUD, no auth/hashing yet

### 4. Load balancing strategy

Only one strategy is implemented today: **consistent hashing** (SHA-256 of the
server address, placed on a sorted ring; `GetServer(key)` hashes the request
key — currently `r.RemoteAddr` — and walks clockwise to the nearest server).
The `Loadbalancer` interface and `balancerFactory` are designed so additional
strategies (round-robin, weighted, least-connections) can be added by
implementing the interface and adding a case to the factory switch.

---

## Getting started

### Prerequisites
- Go 1.25+ (see [go.mod](go.mod))
- A MongoDB instance
- A `.env` file in the project root with:
  ```
  MONGO_URL=mongodb://<user>:<pass>@<host>:<port>
  ```

### Run

```bash
go mod download
go run main.go
```

This will:
- write logs to `./logs/app.log`
- start the internal management API on `http://localhost:8081`
- start the reverse proxy on `http://localhost:8080`

You'll need at least one `Config` document in Mongo first (`POST /createConfig`
with a body like [example_config.txt](example_config.txt)'s create example),
and its ObjectID hardcoded into [main.go](main.go)'s `config.LoadConfig(...)` call.

### Example: create a config, then hit the proxy

```bash
curl -X POST http://localhost:8081/createConfig \
  -H "Content-Type: application/json" \
  -d @example_config.txt   # (trim to just the JSON object)

curl http://localhost:8080/auth/login
```

---

## Known gaps / WIP

These are visible directly from the current code and worth keeping in mind
before relying on this in anything beyond local experimentation:

- **Rate limiter is disconnected** — [security/limiter.go](security/limiter.go) has its own `main()` and `http.ListenAndServe(":8080", ...)`, meaning it's an unused, non-compiling-together standalone prototype (both it and `proxy` claim `:8080`). It needs to become middleware wrapped around the proxy's mux instead.
- **Load balancers don't hot-reload** — `InitLoadBalancer` uses `sync.Once`, so `PUT /updateConfig` updates Mongo and the in-memory `Configs`, but existing hash rings keep serving the old server list until process restart.
- **Startup config ID is hardcoded** — `main.go` loads a specific Mongo `ObjectID` string rather than reading it from env/flag.
- **No path stripping** — the proxy forwards the full incoming path (including the `/auth` prefix) to the backend rather than rewriting it (there's a commented-out `strings.Join(parts[1:], "/")` line).
- **Passwords stored in plaintext** — `User.Password` has no hashing.
- **No SSL/TLS support** — `Server.SSL`/`CertPath`/`KeyPath` exist in the model but aren't used to start a TLS listener.
- **`security_zone`/health-check fields** (`max_fails`, `fail_timeout`, `down`) are parsed into the model but not actively enforced by the load balancer at request time.

---

## Repo layout

```
.
├── main.go                          # entrypoint / wiring
├── logger/logger.go                 # singleton file logger
├── config/configHandler.go          # Mongo → runtime config, hot-reload getters
├── loadBalancer/
│   ├── start.go                     # InitLoadBalancer / GetBalancers
│   ├── balancer.go                  # Loadbalancer interface + factory
│   └── consistent_hasher.go         # consistent-hash ring implementation
├── proxy/reverse_proxy.go           # data-plane reverse proxy handler
├── security/limiter.go              # per-IP token-bucket limiter (standalone, unwired)
├── internal/
│   ├── database/db.go               # Mongo client + collections
│   ├── models/models.go             # BSON/JSON schemas (User, Config, ...)
│   └── controller/
│       ├── controller.go            # Config CRUD handlers
│       ├── usercontroller.go        # User CRUD handlers
│       └── helper.go                # shared Mongo filter / nil-collection guards
└── example_config.txt               # sample create/update Config payloads
```
