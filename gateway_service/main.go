package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	middlewares "gateway_service/middleware"
	"github.com/gorilla/mux"
)

type ServiceConfig struct {
	URL   string
	Paths []string
	Auth  bool
}

var services = map[string]ServiceConfig{
	"blog": {
		URL:   "http://blogs-service:8081",
		Paths: []string{"/blogs"},
		Auth:  true,
	},
	"auth": {
		URL: "http://auth-service:8082",
		Paths: []string{
			"/login",
			"/register",
			"/profile",
			"/validate-admin",
			"/validate-session",
		},
		Auth: false, // Base auth paths don't need authentication
	},
	"events": {
		URL: "http://events-service:8083",
		Paths: []string{
			"/admin/events",
			"/events",
			"/uploads/events", // ✅ Proxy uploads/events to events-service
		},
		Auth: false,
	},
	
	"attractions": {
		URL: "http://attraction-service:8085",
		Paths: []string{
			"/admin/attractions",
			"/attractions",
			"/uploads", // ✅ Correct path for static files
		},
		Auth: false,
	},
}

var pathAuthOverrides = map[string]bool{
	"/admin/events":      true,
	"/admin/attractions": true,
	"/attractions":       true,
}

func main() {
	r := mux.NewRouter()

	// Setup routes for all services
	setupRoutes(r)

	handler := middlewares.CorsMiddleware(r)

	log.Println("Gateway service running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func setupRoutes(r *mux.Router) {
	for _, config := range services {
		for _, path := range config.Paths {
			handler := createProxyHandler(config.URL)

			requiresAuth := config.Auth
			if override, exists := pathAuthOverrides[path]; exists {
				requiresAuth = override
			}

			if requiresAuth {
				handler = middlewares.AuthMiddleware(handler)
			}

			r.PathPrefix(path).Handler(handler)
		}
	}
}

func createProxyHandler(targetServiceURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := url.Parse(targetServiceURL)
		if err != nil {
			http.Error(w, "Invalid target URL", http.StatusInternalServerError)
			return
		}

		// 🔥 Log incoming requests
		log.Printf("Proxying request to: %s%s", target.String(), r.URL.Path)

		proxy := httputil.NewSingleHostReverseProxy(target)

		// 🔑 Fix the redirect issue (if any)
		proxy.ModifyResponse = func(response *http.Response) error {
			if response.StatusCode >= 300 && response.StatusCode < 400 {
				location := response.Header.Get("Location")
				if strings.Contains(location, target.Host) {
					// Replace internal host with gateway URL
					response.Header.Set("Location", strings.Replace(location, target.Host, "localhost:8080", 1))
				}
			}
			return nil
		}

		// Ensure the Host header matches the target
		r.Host = target.Host
		proxy.ServeHTTP(w, r)
	})
}

func serveReverseProxy(w http.ResponseWriter, r *http.Request) {
	target := determineTargetService(r.URL.Path)
	if target == "" {
		http.Error(w, "Service not found", http.StatusBadRequest)
		return
	}

	url, err := url.Parse(target)
	if err != nil {
		log.Printf("Failed to parse target URL %s: %v", target, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", target, err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}

	r.Host = url.Host
	proxy.ServeHTTP(w, r)
}

func determineTargetService(path string) string {
	for _, config := range services {
		for _, servicePath := range config.Paths {
			if path == servicePath || strings.HasPrefix(path, servicePath+"/") {
				return config.URL
			}
		}
	}
	return ""
}
