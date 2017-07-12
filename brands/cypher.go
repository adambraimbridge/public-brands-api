package brands

import (
	"fmt"
	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
	"strings"
)

// Driver interface
type Driver interface {
	Read(id string) (brand Brand, canonicalId string, found bool, err error)
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
	isCanonicalqueryResults := []struct {
		Brand
	}{}

	isCanonicalquery := &neoism.CypherQuery{
		Statement: `
                        MATCH (t:Thing{prefUUID:{uuid}})
			OPTIONAL MATCH (t)<-[:EQUIVALENT_TO]-(x:Thing)
			OPTIONAL MATCH (x)-[:HAS_PARENT]->(p:Thing)
			OPTIONAL MATCH (x)<-[:HAS_PARENT]-(c:Thing)
			RETURN t.prefUUID as id, t.prefLabel as prefLabel, labels(t) as types, t.descriptionXML as descriptionXML, t.strapline as strapline, t.imageUrl as imageUrl,
			collect ( { id: p.uuid, prefId: p.prefUuid, types: labels(p), prefLabel: p.prefLabel } ) AS parentBrands,
			collect ( { id: c.uuid, prefId: c.prefUuid, types: labels(c), prefLabel: c.prefLabel } ) AS childBrands
                `,
		Parameters: neoism.Props{"uuid": uuid},
		Result:     &isCanonicalqueryResults,
	}

	log.Debugf("CypherResult Read Brand for uuid: %s was: %+v", uuid, isCanonicalqueryResults)

	if err := driver.conn.CypherBatch([]*neoism.CypherQuery{isCanonicalquery}); err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v\n", uuid, isCanonicalquery.Statement, err)
		return Brand{}, "", false, fmt.Errorf("Error accessing Brands datastore for uuid: %s", uuid)
	} else if (len(isCanonicalqueryResults)) == 0 {
		canonicalUUid, err := driver.isSourceBrand(uuid)
		return Brand{}, canonicalUUid, false, err
	}

	publicAPITransformation(&isCanonicalqueryResults[0].Brand, driver.env)
	return isCanonicalqueryResults[0].Brand, "", true, nil
}

func (driver CypherDriver) isSourceBrand(uuid string) (string, error) {
	isSourceQueryResults := []struct {
		Brand
	}{}

	isSourceQuery := &neoism.CypherQuery{
		Statement: `
                        MATCH (b:Thing{uuid:{uuid}})-[:EQUIVALENT_TO]->(c:Thing)
			RETURN c.prefUUID as id
                `,
		Parameters: neoism.Props{"uuid": uuid},
		Result:     &isSourceQueryResults,
	}

	if err := driver.conn.CypherBatch([]*neoism.CypherQuery{isSourceQuery}); err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v\n", uuid, isSourceQuery.Statement, err)
		return "", fmt.Errorf("Error accessing Brands datastore for uuid: %s", uuid)
	} else if (len(isSourceQueryResults)) == 0 {
		return "", nil
	}

	return isSourceQueryResults[0].ID, nil
}

func publicAPITransformation(brand *Brand, env string) {
	parents := make([]*Thing, 0)
	children := make([]*Thing, 0)
	types := brand.Types
	if len(brand.Parents) > 0 {
		for _, idx := range brand.Parents {
			parentTypes := idx.Types
			duplicateParent := false
			for _, existingParent := range parents {
				if strings.Contains(existingParent.ID, idx.ID) {
					duplicateParent = true
				}
			}
			if idx.ID != "" && duplicateParent == false {
				newParent := &Thing{ID: mapper.IDURL(idx.ID), Types: mapper.TypeURIs(parentTypes), DirectType: filterToMostSpecificType(parentTypes), APIURL: mapper.APIURL(idx.ID, idx.Types, env), PrefLabel: idx.PrefLabel}
				parents = append(children, newParent)
			}
			duplicateParent = false
		}
		brand.Parents = parents
	} else {
		brand.Parents = parents
	}
	if len(brand.Children) > 0 {
		for _, idx := range brand.Children {
			childTypes := idx.Types
			dubplicateChild := false
			for _, existingChild := range children {
				if strings.Contains(existingChild.ID, idx.ID) {
					dubplicateChild = true
				}
			}
			if idx.ID != "" && dubplicateChild == false {
				newChild := &Thing{ID: mapper.IDURL(idx.ID), Types: mapper.TypeURIs(childTypes), DirectType: filterToMostSpecificType(childTypes), APIURL: mapper.APIURL(idx.ID, idx.Types, env), PrefLabel: idx.PrefLabel}
				children = append(children, newChild)
			}
			dubplicateChild = false
		}
		brand.Children = children
	} else {
		brand.Children = children
	}
	brand.DescriptionXML = brand.DescriptionXML
	brand.ImageURL = brand.ImageURL
	brand.Strapline = brand.Strapline
	brand.APIURL = mapper.APIURL(brand.ID, types, env)
	brand.DirectType = filterToMostSpecificType(types)
	brand.Types = mapper.TypeURIs(types)
	brand.ID = mapper.IDURL(brand.ID)
}

func filterToMostSpecificType(unfilteredTypes []string) string {
	mostSpecificType, _ := mapper.MostSpecificType(unfilteredTypes)
	fullUri := mapper.TypeURIs([]string{mostSpecificType})
	return fullUri[0]
}
