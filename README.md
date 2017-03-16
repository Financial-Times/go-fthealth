This is a Golang implementation of the FT health check standard endpoint as a library that integrates nicely with the standard Go http library.

LATEST VERSION: v2 package contains the latest FT standards implementation, check example below on how to use it.
Note: The v1a package implementation is DEPRECATED, along with the first implementation which is in the main package!!!

Installation:

    go get github.com/Financial-Times/go-fthealth

Example application with a health check:

    package main
    
    import (
        "fmt"
        fthealth "github.com/Financial-Times/go-fthealth/v2"
        "github.com/gorilla/mux"
        "net/http"
    )

    func main() {
        servicesRouter := mux.NewRouter()

        checks := []fthealth.Check{HealthCheck("Some proper neo4j url")}
        healthCheck := &fthealth.HealthCheck{SystemCode: "upp-relations-api", Name: "Relations API", Description: "Retrieves content collection relations from Neo4j", Checks: checks, Parallel: true}
        servicesRouter.HandleFunc("/__health", fthealth.Handler(healthCheck))

        err := http.ListenAndServe(":8080", servicesRouter)
        if err != nil {
            panic(err)
        }
    }

    func HealthCheck(neoURL string) fthealth.Check {
        return fthealth.Check{
            ID:               "check-connectivity-to-neo4j",
            Name:             "Check connectivity to Neo4j",
            Severity:         1,
            BusinessImpact:   "Content collections relations won't be available",
            TechnicalSummary: fmt.Sprintf(`Cannot connect to Neo4j (%v). Check that Neo4j instance is up and running`, neoURL),
            PanicGuide:       "https://dewey.ft.com/upp-relations-api.html",
            Checker:          checker,
        }
    }

    func checker() (string, error) {
        err := func() error { return nil }() // DO A PROPER CHECK HERE
        if err == nil {
            return "Connectivity to Neo4j is ok", err
        }
        return "Error connecting to Neo4j", err
    }
