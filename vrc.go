package main

import (
    "errors"
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
)

type remoteUpstream struct {
    endpoint string
    authType string
    creds []string
}

var remoteUpstreams []remoteUpstream

var modifyResponseEnabled = false
var localUpstream = "http://172.17.0.2:8200"

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
    url, err := url.Parse(targetHost)
    if err != nil {
        return nil, err
    }

    if modifyResponseEnabled {
        log.Print("Starting with request modification enabled.")
        proxy := httputil.NewSingleHostReverseProxy(url)

        originalDirector := proxy.Director
        proxy.Director = func(req *http.Request) {
            originalDirector(req)
            modifyRequest(req)
        }

        proxy.ModifyResponse = modifyResponse()
        proxy.ErrorHandler = errorHandler()
        return proxy, nil

    } else {
        log.Print("Starting with request modification disabled.")
        return httputil.NewSingleHostReverseProxy(url), nil
    }
}


func modifyRequest(req *http.Request) {
    req.Header.Set("X-Proxy", "Simple-Reverse-Proxy")
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
    return func(w http.ResponseWriter, req *http.Request, err error) {
        fmt.Printf("Got error while modifying response: %v \n", err)
        return
    }
}

func modifyResponse() func(*http.Response) error {
    return func(resp *http.Response) error {
        return errors.New("response body is invalid")
    }
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        proxy.ServeHTTP(w, r)
    }
}

func main() {
    // initialize a reverse proxy and pass the actual backend server url here
    proxy, err := NewProxy(localUpstream)
    if err != nil {
        panic(err)
    }

    // handle all requests to your server using the proxy
    http.HandleFunc("/", ProxyRequestHandler(proxy))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
