package brands

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Financial-Times/concepts-rw-neo4j/concepts"
	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

var parentUuid = "d851e146-e889-43f3-8f4c-269da9bb0298"
var secondParentUuid = "a5f3d111-801b-4d98-9b3b-888f57f917ed"
var firstChildUuid = "a806e270-edbc-423f-b8db-d21ae90e06c8"
var secondChildUuid = "d88e2e92-b660-4b6c-a4f0-2184a8fbf051"
var tmeConceptUuid = "bfdc2e18-f50a-4e50-8a04-416779e13f26"
var slConceptUuid = "090987d3-42bd-4479-acb9-279463635093"
var loneNodeUuid = "58231004-22d3-4f86-bf98-13d1390ea06b"

var unfilteredTypes = []string{"http://www.ft.com/ontology/core/Thing", "http://www.ft.com/ontology/concept/Concept", "http://www.ft.com/ontology/classification/Classification", "http://www.ft.com/ontology/product/Brand"}
var nodeLabels = []string{"Thing", "Concept", "Classification", "Brand"}
var db neoutils.NeoConnection

var simpleBrand = Brand{
	Thing:          sourceBrand,
	Strapline:      "Keeping it simple",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
}

var concordedBrand = Brand{
	Thing:          sourceBrand,
	Strapline:      "Keeping it simple",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
	Parents:        []*Thing{&parentBrand},
	Children:       []*Thing{&firstChildBrand},
}

var complexConcordedBrand = Brand{
	Thing:          sourceBrand,
	Strapline:      "Keeping it simple",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
	Parents:        []*Thing{&parentBrand},
	Children:       []*Thing{&firstChildBrand, &secondChildBrand},
}

var noDuplicateParentBrand = Brand{
	Thing:          sourceBrand,
	Strapline:      "Keeping it simple",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
	Parents:        []*Thing{&parentBrand},
}

var multipleParentBrand = Brand{
	Thing:          sourceBrand,
	Strapline:      "Keeping it simple",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
	Parents:        []*Thing{&parentBrand, &secondParent},
}

var sourceBrand = Thing{
	ID:         mapper.IDURL(slConceptUuid),
	PrefLabel:  "The Best Label",
	APIURL:     "http://test.api.ft.com/brands/" + slConceptUuid,
	Types:      unfilteredTypes,
	DirectType: filterToMostSpecificType(nodeLabels),
}

var parentBrand = Thing{
	ID:         mapper.IDURL(parentUuid),
	PrefLabel:  "Parent Brand",
	APIURL:     "http://test.api.ft.com/brands/" + parentUuid,
	Types:      unfilteredTypes,
	DirectType: filterToMostSpecificType(nodeLabels),
}

var secondParent = Thing{
	ID:         mapper.IDURL(secondParentUuid),
	PrefLabel:  "Secondary Parent Brand",
	APIURL:     "http://test.api.ft.com/brands/" + secondParentUuid,
	Types:      unfilteredTypes,
	DirectType: filterToMostSpecificType(nodeLabels),
}

var firstChildBrand = Thing{
	ID:         mapper.IDURL(firstChildUuid),
	PrefLabel:  "First Child Brand",
	APIURL:     "http://test.api.ft.com/brands/" + firstChildUuid,
	Types:      unfilteredTypes,
	DirectType: filterToMostSpecificType(nodeLabels),
}

var secondChildBrand = Thing{
	ID:         mapper.IDURL(secondChildUuid),
	PrefLabel:  "Second Child Brand",
	APIURL:     "http://test.api.ft.com/brands/" + secondChildUuid,
	Types:      unfilteredTypes,
	DirectType: filterToMostSpecificType(nodeLabels),
}

func TestIsSourceBrand(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/parentBrand.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/firstChild.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/dualConcordance.json", assert)

	srv := getBrandDriver(t)
	srv.env = "test"

	type testStruct struct {
		testName     string
		brandUuid    string
		expectedUuid string
	}

	sourceNodeReturnsConcordedId := testStruct{testName: "sourceNodeReturnsConcordedId", brandUuid: tmeConceptUuid, expectedUuid: slConceptUuid}
	loneNodeReturnsEmptyString := testStruct{testName: "parentNodeReturnsEmptyString", brandUuid: loneNodeUuid, expectedUuid: ""}

	testScenarios := []testStruct{sourceNodeReturnsConcordedId, loneNodeReturnsEmptyString}

	for _, scenario := range testScenarios {
		concordedUuid, err := srv.isSourceBrand(scenario.brandUuid)
		assert.NoError(err, "Scenario: "+scenario.testName+" should not return error")
		assert.Equal(scenario.expectedUuid, concordedUuid, "Scenario: "+scenario.testName+" failed. Returned uuid should be "+scenario.expectedUuid)
	}

	defer cleanDB(t)
}

func TestRead_SimpleBrandWithNoParentsOrChildren(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/simpleConcordance.json", assert)
	readAndCompare(simpleBrand, slConceptUuid, 0, 0, t)
	defer cleanDB(t)
}

func TestRead_BrandWithOneParentOneChild(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/parentBrand.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/firstChild.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/dualConcordance.json", assert)
	readAndCompare(concordedBrand, slConceptUuid, 1, 1, t)
	defer cleanDB(t)
}

func TestRead_ComplexBrandWithOneParentAndMultipleChildren(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/parentBrand.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/firstChild.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/secondChild.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/dualConcordance.json", assert)
	readAndCompare(complexConcordedBrand, slConceptUuid, 2, 1, t)
	defer cleanDB(t)
}

func TestRead_BrandWithMultipleParents(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/parentBrand.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/secondParentBrand.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/multipleParent.json", assert)
	readAndCompare(multipleParentBrand, slConceptUuid, 0, 2, t)
	defer cleanDB(t)
}

func TestRead_BrandWithDuplicateParents(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/parentBrand.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/sameParent.json", assert)
	writeJSONToService(brandsWriter, "./fixtures/duplicateParent.json", assert)
	readAndCompare(noDuplicateParentBrand, slConceptUuid, 0, 1, t)
	defer cleanDB(t)
}

func TestRead_ReturnCanonicalIdFromSourceOfCanonicalConcept(t *testing.T) {
	assert := assert.New(t)
	brandsWriter := getConceptsRWDriver(t)
	writeJSONToService(brandsWriter, "./fixtures/dualConcordance.json", assert)
	srv := getBrandDriver(t)

	//Read source node that is not canonical uuid
	brandFromDB, canonicalUuid, found, err := srv.Read(tmeConceptUuid)
	assert.Equal(Brand{}, brandFromDB, "Test failed")
	assert.Equal(slConceptUuid, canonicalUuid, "Test failed")
	assert.False(found, "Test Failed")
	assert.NoError(err, "Test Failed")

	//Read uuid that is not connected to concordance
	secondBrandFromDB, secondCanonicalUuid, secondFound, secondError := srv.Read(parentUuid)
	assert.Equal(Brand{}, secondBrandFromDB, "Test failed")
	assert.Equal("", secondCanonicalUuid, "Test failed")
	assert.False(secondFound, "Test Failed")
	assert.NoError(secondError, "Test Failed")

	defer cleanDB(t)
}

func readAndCompare(expected Brand, uuid string, childCount int, parentCount int, t *testing.T) {
	srv := getBrandDriver(t)
	srv.env = "test"
	brandFromDB, _, found, err := srv.Read(uuid)
	types := brandFromDB.Types
	assert.NotEmpty(t, brandFromDB)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.NotEmpty(t, brandFromDB)
	assert.Equal(t, expected.Thing.ID, brandFromDB.Thing.ID, "Ids not equal")
	assert.Equal(t, expected.Thing.APIURL, brandFromDB.Thing.APIURL, "Api Urls not equal")
	assert.Equal(t, expected.PrefLabel, brandFromDB.PrefLabel, "Pref Labels not equal")
	assert.Equal(t, expected.DescriptionXML, brandFromDB.DescriptionXML, "Description XML not equal")
	assert.Equal(t, expected.Strapline, brandFromDB.Strapline, "Straplines not equal")
	assert.Equal(t, expected.ImageURL, brandFromDB.ImageURL, "Image URLs not equal")
	assert.Equal(t, expected.Types, types, "Types not equal")
	assert.Equal(t, childCount, len(brandFromDB.Children))
	for _, expChild := range expected.Children {
		for _, actChild := range brandFromDB.Children {
			if expChild.ID == actChild.ID {
				assert.Equal(t, expChild.ID, actChild.ID, "Child Ids not equal")
				assert.Equal(t, expChild.PrefLabel, actChild.PrefLabel, "Child Pref Labels not equal")
				assert.Equal(t, expChild.APIURL, actChild.APIURL, "Child Api Urls not equal")
			}
		}
	}
	assert.Equal(t, parentCount, len(brandFromDB.Parents))
	for _, expParent := range expected.Parents {
		for _, actParent := range brandFromDB.Parents {
			if expParent.ID == actParent.ID {
				assert.Equal(t, expParent.ID, actParent.ID, "Parent Ids not equal")
				assert.Equal(t, expParent.PrefLabel, actParent.PrefLabel, "Parent Pref Labels not equal")
				assert.Equal(t, expParent.APIURL, actParent.APIURL, "Parent Api Urls not equal")
			}
		}
	}
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
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}
	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(t, err, "Error setting up connection to %s", url)
	return NewCypherDriver(db, "test")
}

func writeJSONToService(service concepts.Service, pathToJSONFile string, assert *assert.Assertions) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(err)
	dec := json.NewDecoder(f)
	inst, _, errr := service.DecodeJSON(dec)
	assert.NoError(errr)
	errrr := service.Write(inst, "")
	assert.NoError(errrr)
}

func cleanDB(t *testing.T) {
	cleanSourceNodes(t, parentUuid, secondParentUuid, firstChildUuid, secondChildUuid, tmeConceptUuid, slConceptUuid)
	deleteSourceNodes(t, parentUuid, secondParentUuid, firstChildUuid, secondChildUuid, tmeConceptUuid, slConceptUuid)
	cleanConcordedNodes(t, tmeConceptUuid, slConceptUuid, secondChildUuid)
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
