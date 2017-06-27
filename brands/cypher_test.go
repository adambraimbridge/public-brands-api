package brands

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/concepts-rw-neo4j/concepts"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"github.com/jmcvetta/neoism"
	"github.com/Financial-Times/neo-model-utils-go/mapper"
)

var parentUuid = "d851e146-e889-43f3-8f4c-269da9bb0298"
var childUuid = "a806e270-edbc-423f-b8db-d21ae90e06c8"
var tmeConceptUuid = "bfdc2e18-f50a-4e50-8a04-416779e13f26"
var slConceptUuid = "090987d3-42bd-4479-acb9-279463635093"
var unfilteredTypes = []string{"http://www.ft.com/ontology/core/Thing", "http://www.ft.com/ontology/concept/Concept", "http://www.ft.com/ontology/classification/Classification", "http://www.ft.com/ontology/product/Brand"}
var db neoutils.NeoConnection


var concordedBrand = Brand{
	Thing: 		smartlogicSourceBrand,
	Strapline: 	"Keeping it simple",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
	Parents: []*Thing{&parentBrand},
	Children: []*Thing{&childBrand},
}

//var smartlogicSourceBrand = concepts.Concept{
//	UUID: 		slConceptUuid,
//	PrefLabel: 	"The Best Label",
//	Type: 		"http://www.ft.com/ontology/product/Brand",
//	Strapline: 	"Keeping it simple",
//	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
//	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
//	Authority: 	"Smartlogic",
//	AuthorityValue: "123456-SL",
//}

var tmeSourceBrand = concepts.Concept{
	UUID: 		tmeConceptUuid,
	PrefLabel: 	"Tme Label",
	Type: 		"http://www.ft.com/ontology/product/Brand",
	DescriptionXML: "<body>This <i>brand</i> has a parent and a child</body>",
	Authority: 	"TME",
	AuthorityValue: "987654-TME",
	ParentUUIDs: 	[]string{parentUuid},
}
var smartlogicSourceBrand = Thing{
	ID:           mapper.IDURL(slConceptUuid),
	PrefLabel:      "The Best Label",
	APIURL:  "http://test.api.ft.com/brands/" + slConceptUuid,
	Types: filterToMostSpecificType(unfilteredTypes),
}

var parentBrand = Thing{
	ID:           mapper.IDURL(parentUuid),
	PrefLabel:      "Parent Brand",
	APIURL:  "http://test.api.ft.com/brands/" + parentUuid,
	Types: filterToMostSpecificType(unfilteredTypes),
}

var childBrand = Thing{
	ID:           mapper.IDURL(childUuid),
	PrefLabel:      "Child Brand",
	APIURL:  "http://test.api.ft.com/brands/" + childUuid,
	Types: filterToMostSpecificType(unfilteredTypes),
}

//var childBrand = concepts.Concept{
//	UUID:           childUuid,
//	PrefLabel:      "childBrand1",
//	Type: "http://www.ft.com/ontology/product/Brand",
//	ParentUUIDs:     []string{slConceptUuid},
//	Strapline:      "I live in one family",
//	DescriptionXML: "<body>This <i>brand</i> has a parent and valid values for all fields</body>",
//	ImageURL:       "http://media.ft.com/childBrand1.png",
//}

//func TestSimpleBrandWithNoParentsOrChildrenAndOneConcordance(t *testing.T) {
//	assert := assert.New(t)
//	brandsWriter := getConceptsRWDriver(t)
//	writeJSONToService(brandsWriter, "./fixtures/simpleConcordance.json", assert)
//	validConcordedBrand.SourceRepresentations = []concepts.Concept{smartlogicSourceBrand}
//	readAndCompare(&validConcordedBrand, nil, nil, t)
//	defer cleanDB(t)
//}
//
//func TestComplexBrandWithOneParentOneChildAndMultiConcordance(t *testing.T) {
//	assert := assert.New(t)
//	brandsWriter := getConceptsRWDriver(t)
//	writeJSONToService(brandsWriter, "./fixtures/parentWithConcordance.json", assert)
//	writeJSONToService(brandsWriter, "./fixtures/childrenWithConcordance.json", assert)
//	writeJSONToService(brandsWriter, "./fixtures/dualConcordance.json", assert)
//	validConcordedBrand.SourceRepresentations = []concepts.Concept{smartlogicSourceBrand, tmeSourceBrand}
//	readAndCompare(&validConcordedBrand, []*concepts.Concept{&parentBrand}, []*concepts.Concept{&childBrand}, t)
//	//defer cleanDB(t)
//}

func TestComplexBrandWithOneParentOneChildAndMultiConcordance2(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/parentWithConcordance.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/childrenWithConcordance.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/dualConcordance.json", assert)
	readAndCompare2(concordedBrand, slConceptUuid, t)
	//defer cleanDB(t)
}

//func TestSimpleBrandAsParent(t *testing.T) {
//	assert := assert.New(t)
//
//	brandsWriter := getConceptsRWDriver(t)
//	writeJSONToService(brandsWriter, "./fixtures/ParentBrand-d851e146-e889-43f3-8f4c-269da9bb0298.json", assert)
//	writeJSONToService(brandsWriter, "./fixtures/ChildBrand-a806e270-edbc-423f-b8db-d21ae90e06c8.json", assert)
//
//	readAndCompare(&childBrand, []*concepts.Concept{&parentBrand}, nil, t)
//	cleanUp(childBrand.UUID, t)
//	cleanUp(parentBrand.UUID, t)
//}
//
//func TestConnectivityCheck(t *testing.T) {
//	driver := getConceptsRWDriver(t)
//	err := driver.Check()
//	assert.NoError(t, err)
//}

func readAndCompare2(expected Brand, uuid string, t *testing.T) {
	srv := getBrandDriver(t)
	srv.env = "test"
	brand, _, found, err := srv.Read(uuid)
	assert.NotEmpty(t, brand)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.NotEmpty(t, brand)
	assert.Equal(t, expected.Thing.ID, brand.Thing.ID, "Ids not equal")
	assert.Equal(t, expected.Thing.APIURL, brand.Thing.APIURL, "Api Urls not equal")
	assert.Equal(t, expected.PrefLabel, brand.PrefLabel, "Pref Labels not equal")
	assert.Equal(t, expected.DescriptionXML, brand.DescriptionXML, "Description XML not equal")
	assert.Equal(t, expected.Strapline, brand.Strapline, "Straplines not equal")
	assert.Equal(t, expected.ImageURL, brand.ImageURL, "Image URLs not equal")
	//if brand.Children != nil {
	for _, child := range expected.Children {

	}
	assert.Equal(t, expected.Children, brand.Children, "Children not equal")
	//}
	assert.Equal(t, expected.Parents, brand.Parents, "Parents not equal")
}

//func readAndCompare(source *concepts.AggregatedConcept, parents []*concepts.Concept, children []*concepts.Concept, t *testing.T) {
//	srv := getBrandDriver(t)
//	srv.env = "test"
//	brand, _, found, err := srv.Read(source.PrefUUID)
//	expected := makeBrand(source, parents, children, t)
//	fmt.Printf("Resulting brand is %s\n", expected)
//	assert.NoError(t, err)
//	assert.True(t, found)
//	assert.NotEmpty(t, brand)
//	assert.Equal(t, expected.Thing.ID, brand.Thing.ID, "Ids not equal")
//	assert.Equal(t, expected.Thing.APIURL, brand.Thing.APIURL, "Api Urls not equal")
//	assert.Equal(t, expected.PrefLabel, brand.PrefLabel, "Pref Labels not equal")
//	assert.Equal(t, expected.DescriptionXML, brand.DescriptionXML, "Description XML not equal")
//	assert.Equal(t, expected.Strapline, brand.Strapline, "Straplines not equal")
//	assert.Equal(t, expected.ImageURL, brand.ImageURL, "Image URLs not equal")
//	//if brand.Children != nil {
//	assert.Equal(t, expected.Children, brand.Children, "Children not equal")
//	//}
//	assert.Equal(t, expected.Parents, brand.Parents, "Parents not equal")
//}

//func makeBrand(source *concepts.AggregatedConcept, parents []*concepts.Concept, children []*concepts.Concept, t *testing.T) Brand {
//	var brand Brand
//	brand.Thing = makeThing(source, nil, t)
//	brand.PrefLabel = source.PrefLabel
//	brand.Strapline = source.Strapline
//	brand.DescriptionXML = source.DescriptionXML
//	brand.ImageURL = source.ImageURL
//	brand.APIURL = mapper.APIURL(source.PrefUUID, []string{source.Type}, "env")
//	var parentList []*Thing
//	if parents != nil {
//		for _, parent := range parents {
//			newParent := Thing{ID: mapper.IDURL(parent.UUID), PrefLabel: parent.PrefLabel, Types: []string{parent.Type}, APIURL: mapper.APIURL(parent.UUID, []string{parent.Type}, "test")}
//			parentList = append(parentList, &newParent)
//		}
//		brand.Parents = parentList
//	}
//	var childList []*Thing
//	if children != nil {
//		for _, child := range children {
//			newChild := Thing{ID: mapper.IDURL(child.UUID), PrefLabel: child.PrefLabel, Types: []string{child.Type}, APIURL: mapper.APIURL(child.UUID, []string{child.Type}, "test")}
//			childList = append(childList, &newChild)
//		}
//		brand.Children = childList
//	}
//	return brand
//}

func makeThing(uuid string, prefLabel string) *Thing {
	thing := Thing{}
	thing.ID = "http://api.ft.com/things/" + uuid
	thing.APIURL = "http://test.api.ft.com/brands/" + uuid
	thing.Types = []string{"http://www.ft.com/ontology/core/Thing", "http://www.ft.com/ontology/concept/Concept", "http://www.ft.com/ontology/classification/Classification", "http://www.ft.com/ontology/product/Brand"}
	thing.PrefLabel = prefLabel
	return &thing
}

func getConceptsRWDriver(t *testing.T) concepts.Service {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}
	conf := neoutils.DefaultConnectionConfig()
	db, _ = neoutils.Connect(url, conf)
	return concepts.NewConceptService(db)
}

func getBrandDriver(t *testing.T) CypherDriver {
	url := os.Getenv("test")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}
	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(t, err, "Error setting up connection to %s", url)
	return NewCypherDriver(db, "test")
}

func cleanDB(t *testing.T) {
	cleanSourceNodes(t, parentUuid, childUuid, tmeConceptUuid, slConceptUuid)
	deleteSourceNodes(t, parentUuid, childUuid, tmeConceptUuid, slConceptUuid)
	cleanConcordedNodes(t, tmeConceptUuid, slConceptUuid)
}

func writeJSONToService(service concepts.Service, pathToJSONFile string, assert *assert.Assertions) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(err)
	dec := json.NewDecoder(f)
	inst, _, errr := service.DecodeJSON(dec)
	assert.NoError(errr)
	errrr := service.Write(inst)
	assert.NoError(errrr)
}

func deleteSourceNodes(t *testing.T, uuids ...string) {
	qs := make([]*neoism.CypherQuery, len(uuids))
	for i, uuid := range uuids {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf(`
			MATCH (a:Thing {uuid: "%s"})
			OPTIONAL MATCH (a)-[rel]-(i)
			DETACH DELETE rel, i, a`, uuid)}
	}
	err := db.CypherBatch(qs)
	assert.NoError(t, err, "Error executing clean up cypher")
}

func cleanSourceNodes(t *testing.T, uuids ...string) {
	qs := make([]*neoism.CypherQuery, len(uuids))
	for i, uuid := range uuids {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf(`
			MATCH (a:Thing {uuid: "%s"})
			OPTIONAL MATCH (a)-[rel:IDENTIFIES]-(i)
			OPTIONAL MATCH (a)-[hp:HAS_PARENT]-(p)
			DELETE rel, hp, i`, uuid)}
	}
	err := db.CypherBatch(qs)
	assert.NoError(t, err, "Error executing clean up cypher")
}

func cleanConcordedNodes(t *testing.T, uuids ...string) {
	qs := make([]*neoism.CypherQuery, len(uuids))
	for i, uuid := range uuids {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf(`
			MATCH (a:Thing {prefUUID: "%s"})
			OPTIONAL MATCH (a)-[rel]-(i)
			DELETE rel, i, a`, uuid)}
	}
	err := db.CypherBatch(qs)
	assert.NoError(t, err, "Error executing clean up cypher")
}
