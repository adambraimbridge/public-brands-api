package brands

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

// BrandsDriver for cypher queries
var BrandsDriver Driver

// CacheControlHeader is the value to set on http header
var CacheControlHeader string

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

const (
	validUUID     = "([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$"
	thingsApiUrl  = "http://api.ft.com/things/"
	ftThing       = "http://www.ft.com/thing/"
	brandOntology = "http://www.ft.com/ontology/product/Brand"
	queryParams   = "?showRelationship=broader&showRelationship=narrower"
)

type BrandsHandler struct {
	client      httpClient
	conceptsURL string
}

func NewHandler(client httpClient, conceptsURL string) BrandsHandler {
	return BrandsHandler{
		client:      client,
		conceptsURL: conceptsURL,
	}
}

// HealthCheck lightly tests this applications dependencies and returns the results in FT standard format.
func (h *BrandsHandler) HealthCheck() fthealth.TimedHealthCheck {
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
					Checker:          h.Checker,
				},
			},
		},
		Timeout: 10 * time.Second,
	}
}

// Checker does more stuff
func (h *BrandsHandler) Checker() (string, error) {
	err := BrandsDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

// G2GCheck simply checks if we can talk to neo4j
func (h *BrandsHandler) G2GCheck() gtg.Status {
	err := BrandsDriver.CheckConnectivity()
	if err != nil {
		return gtg.Status{GoodToGo: false, Message: "Cannot connect to Neo4J datastore, see healthcheck endpoint for details"}
	}
	return gtg.Status{GoodToGo: true}
}

// MethodNotAllowedHandler does stuff
func (h *BrandsHandler) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

func (h *BrandsHandler) RegisterHandlers(router *mux.Router) {
	logger.Info("Registering handlers")
	mh := handlers.MethodHandler{
		"GET": http.HandlerFunc(h.GetBrand),
	}

	// These paths need to actually be the concept type
	router.Handle("/brands/{uuid}", mh)
}

// GetBrand is the public API
func (h *BrandsHandler) GetBrand(w http.ResponseWriter, r *http.Request) {
	uuidMatcher := regexp.MustCompile(validUUID)
	vars := mux.Vars(r)
	UUID := vars["uuid"]
	transID := transactionidutils.GetTransactionIDFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", CacheControlHeader)

	if UUID == "" || !uuidMatcher.MatchString(UUID) {
		msg := fmt.Sprintf(`uuid '%s' is either missing or invalid`, UUID)
		logger.WithTransactionID(transID).WithUUID(UUID).Error(msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "` + msg + `"}`))
		return
	}

	brand, canonicalUUID, found, err := h.getBrandViaConceptsAPI(UUID, transID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "failed to return brand"}`))
		return
	}

	if found && canonicalUUID != "" && canonicalUUID != UUID {
		redirectURL := strings.Replace(r.RequestURI, UUID, canonicalUUID, 1)
		logger.WithTransactionID(transID).WithUUID(UUID).Debug("serving redirect")
		w.Header().Set("Location", redirectURL)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	if !found {
		msg := fmt.Sprint("brand not found")
		w.WriteHeader(http.StatusNotFound)
		logger.WithTransactionID(transID).WithUUID(UUID).Info(msg)
		w.Write([]byte(`{"message": "` + msg + `"}`))
		return
	}

	log.Debugf("Brand (uuid:%s): %s\n", brand)

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(brand)
	if err != nil {
		msg := fmt.Sprintf("brand: %v could not be marshaled", brand)
		logger.WithError(err).WithTransactionID(transID).WithUUID(UUID).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + msg + `"}`))
	}
}

func (h *BrandsHandler) getBrandViaConceptsAPI(UUID string, transID string) (brand Brand, canonicalUuid string, found bool, err error) {
	logger.WithTransactionID(transID).WithUUID(UUID).Debug("retrieving brand via concepts api")
	mappedBrand := Brand{}
	reqURL := h.conceptsURL + "/concepts/" + UUID + queryParams
	request, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		msg := fmt.Sprintf("failed to create request to %s", reqURL)
		logger.WithError(err).WithUUID(UUID).WithTransactionID(transID).Error(msg)
		return mappedBrand, "", false, err
	}

	request.Header.Set("X-Request-Id", transID)
	resp, err := h.client.Do(request)
	if err != nil {
		msg := fmt.Sprintf("request to %s returned status: %d", reqURL, resp.StatusCode)
		logger.WithError(err).WithUUID(UUID).WithTransactionID(transID).Error(msg)
		return mappedBrand, "", false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return mappedBrand, "", false, nil
	}

	conceptsApiResponse := ConceptApiResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read response body: %v", resp.Body)
		logger.WithError(err).WithUUID(UUID).WithTransactionID(transID).Error(msg)
		return mappedBrand, "", false, err
	}
	if err = json.Unmarshal(body, &conceptsApiResponse); err != nil {
		msg := fmt.Sprintf("failed to unmarshal response body: %v", body)
		logger.WithError(err).WithUUID(UUID).WithTransactionID(transID).Error(msg)
		return mappedBrand, "", false, err
	}

	if conceptsApiResponse.Type != brandOntology {
		logger.WithTransactionID(transID).WithUUID(UUID).Debug("requested concept is not a brand")
		return mappedBrand, "", false, nil
	}

	mappedBrand.ID = convertID(conceptsApiResponse.ID)
	mappedBrand.APIURL = convertApiUrl(conceptsApiResponse.ApiURL)
	mappedBrand.PrefLabel = conceptsApiResponse.PrefLabel
	mappedBrand.Types = mapper.FullTypeHierarchy(conceptsApiResponse.Type)
	mappedBrand.DirectType = conceptsApiResponse.Type
	mappedBrand.ImageURL = conceptsApiResponse.ImageURL
	mappedBrand.DescriptionXML = conceptsApiResponse.DescriptionXML
	mappedBrand.Strapline = conceptsApiResponse.Strapline

	for _, broader := range conceptsApiResponse.Broader {
		if broader.Concept.Type == brandOntology {
			mappedBrand.Parent = convertRelationship(broader)
			break
		}
	}
	var children []Thing
	for _, narrower := range conceptsApiResponse.Narrower {
		children = append(children, *convertRelationship(narrower))
	}
	mappedBrand.Children = children
	return mappedBrand, strings.TrimPrefix(mappedBrand.ID, thingsApiUrl), true, nil
}

func convertRelationship(rc RelatedConcept) *Thing {
	return &Thing{
		ID:         convertID(rc.Concept.ID),
		APIURL:     convertApiUrl(rc.Concept.ApiURL),
		Types:      mapper.FullTypeHierarchy(rc.Concept.Type),
		DirectType: rc.Concept.Type,
		PrefLabel:  rc.Concept.PrefLabel,
	}
}

func convertApiUrl(conceptsApiUrl string) string {
	return strings.Replace(conceptsApiUrl, "concepts", "brands", 1)
}

func convertID(conceptsApiID string) string {
	return strings.Replace(conceptsApiID, ftThing, thingsApiUrl, 1)
}
