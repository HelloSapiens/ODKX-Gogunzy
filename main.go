package main

import (
        "compress/gzip"
        "io"
        "log"
        "net/http"
        "net/url"
        "strings"
)

const backendURL = "http://sync:8080"

func main() {
        http.HandleFunc("/odktables/", proxyHandler)
        http.HandleFunc("/gogunzy-health", healthHandler)

        log.Println("ODKX Go Gzip Proxy listening on :8000...")
        log.Fatal(http.ListenAndServe(":8000", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
        targetURL := backendURL + r.URL.Path
        if r.URL.RawQuery != "" {
                targetURL += "?" + r.URL.RawQuery
        }

        log.Printf("Proxying %s request to %s", r.Method, targetURL)

        var body io.Reader = nil
        cleanHeaders := r.Header.Clone()

        if r.Body != nil {
                body = r.Body
                if cleanHeaders.Get("Content-Encoding") == "gzip" {
                        log.Println("Decompressing gzipped request body...")
                        gz, err := gzip.NewReader(r.Body)
                        if err != nil {
                                http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
                                return
                        }
                        defer gz.Close()
                        body = gz
                        cleanHeaders.Del("Content-Encoding")
                }
        }

        cleanHeaders.Set("Accept-Encoding", "identity")

        req, err := http.NewRequest(r.Method, targetURL, body)
        if err != nil {
                http.Error(w, "Failed to create request", http.StatusInternalServerError)
                return
        }

        req.Header = cleanHeaders
        req.Host = r.Host // preserve original host

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
                http.Error(w, "Upstream request failed", http.StatusBadGateway)
                return
        }
        defer resp.Body.Close()

        for name, values := range resp.Header {
                // Fix potential internal host leaks in redirect responses
                if strings.ToLower(name) == "location" {
                        for _, v := range values {
                                if loc, err := url.Parse(v); err == nil {
                                        if loc.Host == "sync:8080" || loc.Host == "sync" {
                                                loc.Scheme = "https"
                                                loc.Host = r.Host
                                                v = loc.String()
                                        }
                                }
                                w.Header().Add(name, v)
                        }
                } else {
                        for _, v := range values {
                                w.Header().Add(name, v)
                        }
                }
        }

        w.WriteHeader(resp.StatusCode)
        io.Copy(w, resp.Body)
}