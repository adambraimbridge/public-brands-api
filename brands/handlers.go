package brands

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"

	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/gorilla/mux"
)

// BrandsDriver for cypher queries
var BrandsDriver Driver

// CacheControlHeader is the value to set on http header
var CacheControlHeader string

// HealthCheck does something
func HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to Public Brand api requests",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Neo4j a instance",
		Checker:          Checker,
	}
}

// Checker does more stuff
func Checker() (string, error) {
	err := BrandsDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

// MethodNotAllowedHandler does stuff
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

// Ping says pong
func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

// BuildInfo - This is a stop gap and will be added to when we can define what we should display here
func BuildInfo(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "build-info")
}

// GetBrand is the public API
func GetBrand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if uuid == "" {
		http.Error(w, "uuid required", http.StatusBadRequest)
		return
	}
	brand, found, err := BrandsDriver.Read(uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Brand not found."}`))
		return
	}
	log.Debugf("Brand (uuid:%s): %s\n", brand)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(brand)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Organisation could not be marshelled, err=` + err.Error() + `"}`))
	}
}
