package proxy

import(
    "net/url"
    "net/http"
    "net/http/httputil"
)

func create_proxy() *httputil.ReverseProxy {

	target,err := url.Parse("http://localhost:9000")
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	return proxy
}

func serve(){
	proxy := create_proxy()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
	http.ListenAndServe(":8080", nil)
}