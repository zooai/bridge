// Copyright (C) 2026, Zoo Labs Foundation. All rights reserved.
// Zoo Bridge -- reverse proxy for the Lux Bridge server with Zoo branding.
//
// The upstream Lux Bridge is a TypeScript Express server (ghcr.io/luxfi/bridge-server).
// This binary sits in front of it as a Go reverse proxy, adding:
//   - Zoo-specific CORS origins
//   - Brand headers (X-Brand-Name: Zoo)
//   - Health/readiness endpoints for K8s probes
//   - Graceful shutdown
//
// In K8s, the bridge-server runs as a sidecar or separate service.
// This proxy forwards all traffic to it.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		fmt.Printf("zoo-bridge %s\n", version)
		os.Exit(0)
	}

	listenAddr := envOr("BRIDGE_LISTEN", ":8080")
	upstreamURL := envOr("BRIDGE_UPSTREAM_URL", "http://127.0.0.1:5000")
	brandName := envOr("BRIDGE_BRAND_NAME", "Zoo")
	corsOrigins := envOr("BRIDGE_CORS_ORIGINS", "https://bridge.zoo.ngo,https://zoo.ngo")

	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		log.Fatalf("invalid BRIDGE_UPSTREAM_URL %q: %v", upstreamURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(upstream)

	// Preserve the original director, add brand header.
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Header.Set("X-Brand-Name", brandName)
		req.Host = upstream.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %s %s: %v", r.Method, r.URL.Path, err)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `{"error":"upstream unavailable","service":"zoo-bridge"}`)
	}

	allowedOrigins := make(map[string]bool)
	for _, o := range strings.Split(corsOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowedOrigins[o] = true
		}
	}

	mux := http.NewServeMux()

	// Health check -- does not hit upstream.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","service":"zoo-bridge","version":"%s"}`, version)
	})

	// Readiness -- pings upstream health endpoint.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL+"/health", nil)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"error","error":"%s"}`, err.Error())
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","upstream":"%s","error":"%s"}`, upstreamURL, err.Error())
			return
		}
		resp.Body.Close()
		if resp.StatusCode >= 500 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","upstream_status":%d}`, resp.StatusCode)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ready","service":"zoo-bridge","version":"%s"}`, version)
	})

	// All other paths proxy to upstream.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		proxy.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown on SIGTERM/SIGINT.
	done := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		log.Printf("received %v, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
		close(done)
	}()

	log.Printf("zoo-bridge %s listening on %s, proxying to %s", version, listenAddr, upstreamURL)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen error: %v", err)
	}
	<-done
	log.Println("zoo-bridge stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
