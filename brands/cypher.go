package brands

import (
	"fmt"

	"errors"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	log "github.com/sirupsen/logrus"
)

// Driver interface
type Driver interface {
	Read(id string) (brand Brand, canonicalUuid string, found bool, err error)
	CheckConnectivity() error
}

// CypherDriver struct
type CypherDriver struct {
	conn neoutils.NeoConnection
	env  string
}

//NewCypherDriver instantiate driver
func NewCypherDriver(conn neoutils.NeoConnection, env string) CypherDriver {
	return CypherDriver{conn, env}
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (driver CypherDriver) CheckConnectivity() error {
	return neoutils.Check(driver.conn)
}

func (driver CypherDriver) Read(uuid string) (Brand, string, bool, error) {
	results := []NeoBrand{}

	query := &neoism.CypherQuery{
		Statement: `
            MATCH (brand:Brand{uuid:{uuid}})-[:EQUIVALENT_TO]->(canonicalBrand:Brand)
			OPTIONAL MATCH (canonicalBrand)<-[:EQUIVALENT_TO]-(leafBrand:Brand)
			OPTIONAL MATCH (leafBrand)-[:HAS_PARENT]->(parent:Thing)-[:EQUIVALENT_TO]->(canonicalParent:Thing)
			OPTIONAL MATCH (leafBrand)<-[:HAS_PARENT]-(children:Thing)-[:EQUIVALENT_TO]->(canonicalChildren:Thing)
			RETURN canonicalBrand.prefUUID as ID, labels(canonicalBrand) as types, canonicalBrand.prefLabel as prefLabel,
				canonicalBrand.descriptionXML as descriptionXML, canonicalBrand.strapline as strapline, canonicalBrand.imageUrl as imageUrl,
				leafBrand.authority as authority, {id: canonicalParent.prefUUID, types: labels(canonicalParent), prefLabel: canonicalParent.prefLabel} as parent,
				collect({id: canonicalChildren.prefUUID, types: labels(canonicalChildren), prefLabel: canonicalChildren.prefLabel}) as children`,
		Parameters: neoism.Props{"uuid": uuid},
		Result:     &results,
	}

	log.WithFields(log.Fields{"UUID": uuid, "Results": results, "Query": query}).Debug("CypherResult for New Concordance Model Read Brand")

	err := driver.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "Query": query.Statement}).Errorf("Error looking up canonical Brand with neoism")
		return Brand{}, "", false, fmt.Errorf("Error accessing Brands datastore for uuid: %s", uuid)
	}

	// New model returned results
	if len(results) > 0 {
		publicBrand, UUID := getPublicBrand(results, driver.env)
		return publicBrand, UUID, true, nil
	} else {
		return Brand{}, "", false, nil
	}
}

func getThingFromNeoThing(thing NeoThing, env string) (Thing, error) {
	if thing.ID != "" {
		return Thing{
			mapper.IDURL(thing.ID),
			mapper.APIURL(thing.ID, thing.Types, env),
			mapper.TypeURIs(thing.Types),
			filterToMostSpecificType(thing.Types),
			thing.PrefLabel,
		}, nil
	}
	return Thing{}, errors.New("No thing found")
}

func getPublicBrand(brands []NeoBrand, env string) (Brand, string) {
	for _, brand := range brands {
		if brand.Authority == "Smartlogic" {
			return publicAPITransformation(brand, env), brand.ID
		}
	}
	return publicAPITransformation(brands[0], env), brands[0].ID
}

func publicAPITransformation(brand NeoBrand, env string) Brand {
	var publicBrand Brand

	types := brand.Types

	if parent, err := getThingFromNeoThing(brand.Parent, env); err == nil {
		publicBrand.Parent = &parent
	}

	if len(brand.Children) > 0 {
		children := map[string]Thing{}
		for _, child := range brand.Children {
			if child.ID == "" {
				continue
			}

			if ch, err := getThingFromNeoThing(child, env); err == nil {
				children[child.ID] = ch
			}
		}
		publicBrand.Children = []Thing{}
		for _, v := range children {
			publicBrand.Children = append(publicBrand.Children, v)
		}
	}

	publicBrand.ID = mapper.IDURL(brand.ID)
	publicBrand.APIURL = mapper.APIURL(brand.ID, types, env)
	publicBrand.Types = mapper.TypeURIs(types)
	publicBrand.DirectType = filterToMostSpecificType(types)
	publicBrand.PrefLabel = brand.PrefLabel
	publicBrand.DescriptionXML = brand.DescriptionXML
	publicBrand.ImageURL = brand.ImageURL
	publicBrand.Strapline = brand.Strapline

	return publicBrand
}

func filterToMostSpecificType(unfilteredTypes []string) string {
	mostSpecificType, _ := mapper.MostSpecificType(unfilteredTypes)
	fullUri := mapper.TypeURIs([]string{mostSpecificType})
	return fullUri[0]
}
