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
	newConcordanceModelResults := []struct {
		Brand
	}{}

	newConcordanceModelQuery := &neoism.CypherQuery{
		Statement: `
            MATCH (brand:Brand{uuid:{uuid}})-[:EQUIVALENT_TO]->(canonicalNode:Brand)
			OPTIONAL MATCH (canonicalNode)<-[:EQUIVALENT_TO]-(slBrand:Brand{authority:"Smartlogic"})
			OPTIONAL MATCH (slBrand)-[:HAS_PARENT]->(parent:Thing)-[:EQUIVALENT_TO]->(canonicalParent:Thing)
			OPTIONAL MATCH (slBrand)<-[:HAS_PARENT]-(children:Thing)-[:EQUIVALENT_TO]->(canonicalChildren:Thing)
			RETURN canonicalNode.prefUUID as id, canonicalNode.prefLabel as prefLabel, labels(canonicalNode) as types, canonicalNode.descriptionXML as descriptionXML, canonicalNode.strapline as strapline, canonicalNode.imageUrl as _imageUrl,
			{id: canonicalParent.prefUUID, types: labels(canonicalParent), prefLabel: canonicalParent.prefLabel} as parentBrand,
			collect({id: canonicalChildren.prefUUID, types: labels(canonicalChildren), prefLabel: canonicalChildren.prefLabel}) as childBrands`,
		Parameters: neoism.Props{"uuid": uuid},
		Result:     &newConcordanceModelResults,
	}

	log.WithFields(log.Fields{"UUID": uuid, "Results": newConcordanceModelResults, "Query": newConcordanceModelQuery}).Debug("CypherResult for New Concordance Model Read Brand")

	err := driver.conn.CypherBatch([]*neoism.CypherQuery{newConcordanceModelQuery})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "Query": newConcordanceModelQuery.Statement}).Errorf("Error looking up canonical Brand with neoism")
		return Brand{}, "", false, fmt.Errorf("Error accessing Brands datastore for uuid: %s", uuid)
	}

	// New model returned results
	if (len(newConcordanceModelResults) > 0) {
		canonicalUUID := newConcordanceModelResults[0].Brand.ID
		publicAPITransformation(&newConcordanceModelResults[0].Brand, driver.env)
		return newConcordanceModelResults[0].Brand, canonicalUUID, true, nil
	} else {
		return driver.oldConcordanceModel(uuid)
	}
}

func (driver CypherDriver) oldConcordanceModel(uuid string) (Brand, string, bool, error) {
	oldConcordanceModelResults := []struct {
		Brand
	}{}

	oldConcordanceModelQuery := &neoism.CypherQuery{
		Statement: `
                   MATCH (upp:UPPIdentifier{value:{uuid}})-[:IDENTIFIES]->(b:Brand)
                   OPTIONAL MATCH (b)-[:HAS_PARENT]->(p:Thing)
                   OPTIONAL MATCH (b)<-[:HAS_PARENT]-(c:Thing)
                   RETURN  b.uuid as id, labels(b) as types, b.prefLabel as prefLabel,
                           b.description as description, b.descriptionXML as descriptionXML,
                           b.strapline as strapline, b.imageUrl as _imageUrl,
                           { id: p.uuid, types: labels(p), prefLabel: p.prefLabel } AS parentBrand,
                           collect ({ id: c.uuid, types: labels(c), prefLabel: c.prefLabel }) AS childBrands
                `,
		Parameters: neoism.Props{"uuid": uuid},
		Result:     &oldConcordanceModelResults,
	}

	log.WithFields(log.Fields{"UUID": uuid, "Results": oldConcordanceModelResults, "Query": oldConcordanceModelQuery}).Debug("CypherResult for New Concordance Model Read Brand")

	if err := driver.conn.CypherBatch([]*neoism.CypherQuery{oldConcordanceModelQuery}); err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "Query": oldConcordanceModelQuery.Statement}).Errorf("Error looking up source Brand with neoism")
		return Brand{}, "", false, fmt.Errorf("Error accessing Brands datastore for uuid: %s", uuid)
	} else if (len(oldConcordanceModelResults)) == 0 {
		return Brand{}, "", false, nil
	}
	canonicalUUID := oldConcordanceModelResults[0].Brand.ID
	publicAPITransformation(&oldConcordanceModelResults[0].Brand, driver.env)
	return oldConcordanceModelResults[0].Brand, canonicalUUID, true, nil
}

func publicAPITransformation(brand *Brand, env string) {

	children := make([]*Thing, 0)
	types := brand.Types
	
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

	}

	if brand.Parent != nil && len(brand.Parent.Types) > 0{
		parentTypes := brand.Parent.Types
		brand.Parent.APIURL = mapper.APIURL(brand.Parent.ID, types, env)
		brand.Parent.ID = mapper.IDURL(brand.Parent.ID)
	    brand.Parent.Types =  mapper.TypeURIs(parentTypes)
		brand.Parent.DirectType = filterToMostSpecificType(parentTypes)
	}

	brand.Children = children
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
