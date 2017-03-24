[![Circle CI](https://circleci.com/gh/Financial-Times/go-fthealth.svg?style=shield)](https://circleci.com/gh/Financial-Times/go-fthealth)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/go-fthealth)](https://goreportcard.com/report/github.com/Financial-Times/go-fthealth) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/go-fthealth/badge.svg)](https://coveralls.io/github/Financial-Times/go-fthealth)

This is a Golang implementation of the FT health check standard endpoint as a library that integrates nicely with the standard Go http library.

LATEST VERSION: v1_1 (parity with v1.1 approved in FT health standard document) package contains the latest FT standards implementation, check example below on how to use it.

Note: The v1a package implementation is DEPRECATED, along with the first implementation which is in the main package!!!

Installation:

    go get github.com/Financial-Times/go-fthealth

Example application with a health check:

    package main
    
    import (
        "fmt"
        fthealth "github.com/Financial-Times/go-fthealth/v1_1"
        "github.com/gorilla/mux"
        "net/http"
    )

    func main() {
        servicesRouter := mux.NewRouter()

        checks := []fthealth.Check{MyCheck("Some proper neo4j url")}
        healthCheck := &fthealth.HealthCheck{SystemCode: "upp-relations-api", Name: "Relations API", Description: "Retrieves content collection relations from Neo4j", Checks: checks}
        servicesRouter.HandleFunc("/__health", fthealth.Handler(healthCheck))

        err := http.ListenAndServe(":8080", servicesRouter)
        if err != nil {
            panic(err)
        }
    }

    func MyCheck(neoURL string) fthealth.Check {
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
