package main

import (
    "crypto/tls"
    "github.com/gorilla/mux"
    "github.com/justinas/alice"
    "golang.org/x/crypto/acme/autocert"
    "log"
    "net/http"
    "time"
)


func main() {
    errorChain := alice.New(loggerHandler, recoverHandler)

    var r = mux.NewRouter()
    r.HandleFunc("/", rootHandler)
    r.HandleFunc("/welcome", rootHandler).Name("welcome")
    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

    http.Handle("/", errorChain.Then(r))

    m := autocert.Manager{
        Prompt:     autocert.AcceptTOS,
        HostPolicy: autocert.HostWhitelist("www.mywebsite.com"),
        Cache:      autocert.DirCache("/home/letsencrypt/"),
    }

    server := &http.Server{
        Addr: ":443",
        TLSConfig: &tls.Config{
            GetCertificate: m.GetCertificate,
        },
    }

    log.Printf("Service UP\n")

    err := server.ListenAndServeTLS("", "")
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}


func rootHandler(w http.ResponseWriter, r *http.Request) {
    // render("index.html", w, r)
}


func loggerHandler(h http.Handler) http.Handler {

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        h.ServeHTTP(w, r)
        log.Printf("<< %s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func recoverHandler(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %+v", err)
                http.Error(w, http.StatusText(500), 500)
            }
        }()

        next.ServeHTTP(w, r)
    }

    return http.HandlerFunc(fn)
}
