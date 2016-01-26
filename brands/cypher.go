package brands

import (
	"errors"
	"fmt"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type Driver interface {
	Read(id string) (brand Brand, found bool, err error)
	CheckConnectivity() error
}

// CypherDriver struct
type CypherDriver struct {
	db *neoism.Database
}

//NewCypherDriver instantiate driver
func NewCypherDriver(db *neoism.Database) CypherDriver {
	return CypherDriver{db}
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (pcw CypherDriver) CheckConnectivity() error {
	results := []struct {
		ID int
	}{}
	query := &neoism.CypherQuery{
		Statement: "MATCH (x) RETURN ID(x) LIMIT 1",
		Result:    &results,
	}
	err := pcw.db.Cypher(query)
	log.Debugf("CheckConnectivity results:%+v  err: %+v", results, err)
	return err
}

func (pcw CypherDriver) Read(uuid string) (brand Brand, found bool, err error) {
	results := []struct {
		Brand
	}{}
	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (b:Brand{uuid:{uuid}})
                        OPTIONAL MATCH (b)-[:HAS_PARENT]->(p:Thing)
                        RETURN b.uuid as id, labels(b) as types, b.prefLabel as prefLabel,
                                b.description as description, b.descriptionXML as descriptionXML,
                                b.strapline as strapline, b.imageUrl as _imageUrl,
                                { id: p.uuid, types: labels(p), prefLabel: p.prefLabel,
                                  description: p.description, descriptionXML: p. descriptionXML,
                                  strapline: p.strapline, _imageUrl: p.imageUrl
                                } as parentBrand
                `,
		Parameters: neoism.Props{"uuid": uuid},
		Result:     &results,
	}
	err = pcw.db.Cypher(query)
	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v\n", uuid, query.Statement, err)
		return Brand{}, false, fmt.Errorf("Error accessing Person datastore for uuid: %s", uuid)
	}
	log.Debugf("CypherResult ReadPeople for uuid: %s was: %+v", uuid, results)
	if (len(results)) == 0 {
		return Brand{}, false, nil
	} else if len(results) != 1 {
		errMsg := fmt.Sprintf("Multiple people found with the same uuid:%s !", uuid)
		log.Error(errMsg)
		return Brand{}, true, errors.New(errMsg)
	}
	publicAPITransformation(&results[0].Brand)
	log.Debugf("Returning %v", results[0].Brand)
	return results[0].Brand, true, nil
}

func publicAPITransformation(brand *Brand) {
	if brand.Parent.ID != "" {
		brand.Parent.APIURL = mapper.APIURL(brand.Parent.ID, brand.Parent.Types)
		brand.Parent.Types = mapper.TypeURIs(brand.Parent.Types)
		brand.Parent.ID = mapper.IDURL(brand.Parent.ID)
	} else {
		brand.Parent = nil
	}
	brand.APIURL = mapper.APIURL(brand.ID, brand.Types)
	brand.Types = mapper.TypeURIs(brand.Types)
	brand.ID = mapper.IDURL(brand.ID)
}
