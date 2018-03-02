package brands

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// BrandsDriver for cypher queries
var BrandsDriver Driver

// CacheControlHeader is the value to set on http header
var CacheControlHeader string

const validUUID = "([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$"

// HealthCheck lightly tests this applications dependencies and returns the results in FT standard format.
func HealthCheck() fthealth.TimedHealthCheck {
	return fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			Name:        "Public Brands API",
			SystemCode:  "public-brands-api",
			Description: "A public RESTful API for accessing Brands in neo4j",
			Checks: []fthealth.Check{
				{
					BusinessImpact:   "Unable to respond to Public Brands API requests",
					Name:             "Check connectivity to Neo4j",
					PanicGuide:       "https://dewey.in.ft.com/view/system/public-brands-api",
					Severity:         2,
					TechnicalSummary: "Cannot connect to Neo4j a instance",
					Checker:          Checker,
				},
			},
		},
		Timeout: 10 * time.Second,
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

// G2GCheck simply checks if we can talk to neo4j
func G2GCheck() gtg.Status {
	err := BrandsDriver.CheckConnectivity()
	if err != nil {
		return gtg.Status{GoodToGo: false, Message: "Cannot connect to Neo4J datastore, see healthcheck endpoint for details"}
	}
	return gtg.Status{GoodToGo: true}
}

// MethodNotAllowedHandler does stuff
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
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
	brand, canonicalUUID, found, err := BrandsDriver.Read(uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	if found && canonicalUUID != "" && canonicalUUID != uuid {
		validRegexp := regexp.MustCompile(validUUID)
		canonicalUUID := validRegexp.FindString(canonicalUUID)
		redirectURL := strings.Replace(r.RequestURI, uuid, canonicalUUID, 1)
		w.Header().Set("Location", redirectURL)
		w.WriteHeader(http.StatusMovedPermanently)
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
