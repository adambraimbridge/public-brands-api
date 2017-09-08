package brands

import (
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

var unfilteredTypes = []string{"http://www.ft.com/ontology/core/Thing", "http://www.ft.com/ontology/concept/Concept", "http://www.ft.com/ontology/classification/Classification", "http://www.ft.com/ontology/product/Brand"}
var nodeLabels = []string{"Thing", "Concept", "Classification", "Brand"}
var db neoutils.NeoConnection

// Test 1 - Simple Smartlogic Brand with no parent nor child, representing a brand new brand added to Smartlogic
var simpleSLBrandUUID = "67b4d5ba-4a7b-4d7e-9f34-c00afab865b6"
var simpleAPIOutput = Brand{
	Thing: Thing{mapper.IDURL(simpleSLBrandUUID), "http://test.api.ft.com/brands/" + simpleSLBrandUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Simple Brand New Brand"},
	Strapline:      "Keeping it simple",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
}
var simpleConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: simpleSLBrandUUID, PrefLabel: "Simple Brand New Brand", Type: "Brand",
	Strapline: "Keeping it simple", DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL: "http://media.ft.com/validSmartlogicBrand.png", SourceRepresentations: []concepts.Concept{{
		UUID:           simpleSLBrandUUID,
		PrefLabel:      "Concept PrefLabel",
		Type:           "Brand",
		Authority:      "Smartlogic",
		AuthorityValue: simpleSLBrandUUID,
	}},
}

// Test 2a and 2b - Simple Smartlogic Brand only one, representing a brand new brand added to Smartlogic with a parent
var simpleSLBrandWithParentUUID = "0049efc0-2849-4cab-babf-ca1319fb69d4"
var simpleSLBrandParentsUUID = "54b85a21-e605-41db-9d1d-2bf66193f4ff"

var simpleWithParentAPIOutput = Brand{
	Thing: Thing{mapper.IDURL(simpleSLBrandWithParentUUID), "http://test.api.ft.com/brands/" + simpleSLBrandWithParentUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Simple Brand New Brand With Parent"},
	Strapline:      "Keeping it simple but I have a parent",
	DescriptionXML: "<body>This <i>brand</i> has parent no concordance</body>",
	ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
	Parent: &Thing{mapper.IDURL(simpleSLBrandParentsUUID), "http://test.api.ft.com/brands/" + simpleSLBrandParentsUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Simple Brand New Brands Parent"},
}

var simpleWithChildAPIOutput = Brand{
	Thing: Thing{mapper.IDURL(simpleSLBrandParentsUUID), "http://test.api.ft.com/brands/" + simpleSLBrandParentsUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Simple Brand New Brands Parent"},
	Children: []Thing{Thing{mapper.IDURL(simpleSLBrandParentsUUID), "http://test.api.ft.com/brands/" + simpleSLBrandParentsUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Simple Brand New Brand With Parent"}},
}

var simpleWithParentConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: simpleSLBrandWithParentUUID, PrefLabel: "Simple Brand New Brand With Parent", Type: "Brand",
	Strapline: "Keeping it simple but I have a parent", DescriptionXML: "<body>This <i>brand</i> has parent no concordance</body>",
	ImageURL: "http://media.ft.com/validSmartlogicBrand.png", SourceRepresentations: []concepts.Concept{{
		UUID:           simpleSLBrandWithParentUUID,
		PrefLabel:      "Simple Brand New Brand With Parent",
		Type:           "Brand",
		Authority:      "Smartlogic",
		Strapline:      "Keeping it simple but I have a parent",
		DescriptionXML: "<body>This <i>brand</i> has parent no concordance</body>",
		AuthorityValue: simpleSLBrandWithParentUUID,
		ParentUUIDs:    []string{simpleSLBrandParentsUUID},
		ImageURL:       "http://media.ft.com/validSmartlogicBrand.png",
	}},
}

var simpleWithParentParentConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: simpleSLBrandParentsUUID, PrefLabel: "Simple Brand New Brands Parent", Type: "Brand", SourceRepresentations: []concepts.Concept{{
		UUID:           simpleSLBrandParentsUUID,
		PrefLabel:      "Simple Brand New Brands Parent",
		Type:           "Brand",
		Authority:      "Smartlogic",
		AuthorityValue: simpleSLBrandParentsUUID,
	}},
}

// Test 3 - Concordance with a TME brand but neither having a parent
var concordedBrandWithNoParentUUID = "6cfcf82f-9b58-48e2-99c9-4fc73ff083a8"
var concordedBrandWithNoParentTMEUUID = "f5533f30-f153-4578-af8d-cf0b17ccb869"
var concordedBrandWithNoParentAPIOutput = Brand{
	Thing: Thing{mapper.IDURL(concordedBrandWithNoParentUUID), "http://test.api.ft.com/brands/" + concordedBrandWithNoParentUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Smarto Logico Concorded with one TME no Parent"},
	Strapline: "Keeping it simple but loving TME", ImageURL: "http://media.ft.com/validSmartlogicBrand.png",
}

var concordedBrandWithNoParentConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: concordedBrandWithNoParentUUID, PrefLabel: "Smarto Logico Concorded with one TME no Parent", Type: "Brand",
	ImageURL: "http://media.ft.com/validSmartlogicBrand.png", Strapline: "Keeping it simple but loving TME", SourceRepresentations: []concepts.Concept{{
		UUID:           concordedBrandWithNoParentUUID,
		PrefLabel:      "Smarto Logico Concorded with one TME no Parent",
		Type:           "Brand",
		Authority:      "Smartlogic",
		AuthorityValue: concordedBrandWithNoParentUUID,
		Strapline:      "Keeping it simple but loving TME",
	}, {
		UUID:           concordedBrandWithNoParentTMEUUID,
		PrefLabel:      "Concorded with one TME no Parent",
		Type:           "Brand",
		Authority:      "TME",
		AuthorityValue: "1234-XYZ",
	}},
}

// Test 4 - Concordance with a TME brand both having parents and preferring Smartlogic
var concordedBrandWithParentsUUID = "bf9bd9f4-8a3c-4adc-81f0-461911bbbf5f"
var smartLogicParentUUID = "d9947333-53e9-46a5-a7f4-4190197b621c"
var TMEParentUIUD = "12c50d6c-fd17-4d06-a6e2-a3bd6afbe67f"
var concordedBrandWithParentsTMEUUID = "3a5f50a3-f1d9-434a-a62f-c622fa6a20c1"
var concordedBrandBothWithParentsAPIOutput = Brand{
	Thing: Thing{mapper.IDURL(concordedBrandWithParentsUUID), "http://test.api.ft.com/brands/" + concordedBrandWithParentsUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Smarto Logico Concorded with one TME with Parents"},
	Strapline: "Loving TME and all the parents", ImageURL: "http://media.ft.com/validSmartlogicBrand.png",
	Parent: &Thing{mapper.IDURL(smartLogicParentUUID), "http://test.api.ft.com/brands/" + smartLogicParentUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Parent SL Concept PrefLabel"},
}

var concordedBrandBothWithParentsConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: concordedBrandWithParentsUUID, PrefLabel: "Smarto Logico Concorded with one TME with Parents", Type: "Brand",
	ImageURL: "http://media.ft.com/validSmartlogicBrand.png", Strapline: "Loving TME and all the parents", SourceRepresentations: []concepts.Concept{{
		UUID:           concordedBrandWithParentsUUID,
		PrefLabel:      "Smarto Logico Concorded with one TME with Parents",
		Type:           "Brand",
		Authority:      "Smartlogic",
		AuthorityValue: concordedBrandWithParentsUUID,
		Strapline:      "Loving TME and all the parents",
		ParentUUIDs:    []string{smartLogicParentUUID},
	}, {
		UUID:           concordedBrandWithParentsTMEUUID,
		PrefLabel:      "Concorded with one TME with Parent",
		Type:           "Brand",
		Authority:      "TME",
		AuthorityValue: "427845-XYZ",
		ParentUUIDs:    []string{TMEParentUIUD},
	}},
}

var concordedBrandBothWithParentsSLParentConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: smartLogicParentUUID, PrefLabel: "Parent SL Concept PrefLabel", Type: "Brand", SourceRepresentations: []concepts.Concept{{
		UUID:           smartLogicParentUUID,
		PrefLabel:      "Parent SL Concept PrefLabel",
		Type:           "Brand",
		Authority:      "Smartlogic",
		AuthorityValue: smartLogicParentUUID,
	}},
}

var concordedBrandBothWithParentsTMEParentConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: TMEParentUIUD, PrefLabel: "Parent TME Concept PrefLabel", Type: "Brand", SourceRepresentations: []concepts.Concept{{
		UUID:           TMEParentUIUD,
		PrefLabel:      "Parent TME Concept PrefLabel",
		Type:           "Brand",
		Authority:      "TME",
		AuthorityValue: TMEParentUIUD,
	}},
}

// Test 5 - Concordance with a TME brand having a parent but Smartlogic has none so none is returned
var concordedBrandWithTMEParentOnlyUUID = "57d759c5-c786-47b2-b94b-f779625e7310"
var concordedBrandWithTMEParentOnlyTMEUUID = "cfefc351-3288-4349-b80f-68c50996944b"
var concordedBrandWithTMEParentOnlyTMEParentUIUD = "fc16c4b0-a7c8-4454-805f-83a2fc975e5d"
var concordedBrandWithTMEParentOnlyAPIOutput = Brand{
	Thing: Thing{mapper.IDURL(concordedBrandWithTMEParentOnlyUUID), "http://test.api.ft.com/brands/" + concordedBrandWithTMEParentOnlyUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Smarto Logico Concorded with one TME with Parents"},
}

var concordedBrandWithTMEParentOnlyConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: concordedBrandWithTMEParentOnlyUUID, PrefLabel: "Smarto Logico Concorded with one TME with Parents", Type: "Brand", SourceRepresentations: []concepts.Concept{{
		UUID:           concordedBrandWithTMEParentOnlyUUID,
		PrefLabel:      "Smarto Logico Concorded with one TME with Parents",
		Type:           "Brand",
		Authority:      "Smartlogic",
		AuthorityValue: concordedBrandWithTMEParentOnlyUUID,
	}, {
		UUID:           concordedBrandWithTMEParentOnlyTMEUUID,
		PrefLabel:      "Concorded with one TME with Parent",
		Type:           "Brand",
		Authority:      "TME",
		AuthorityValue: "54321-XYZ",
		ParentUUIDs:    []string{concordedBrandWithTMEParentOnlyTMEParentUIUD},
	}},
}

var concordedBrandWithTMEParentOnlyParentConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: concordedBrandWithTMEParentOnlyTMEParentUIUD, PrefLabel: "Concept PrefLabel", Type: "Brand", SourceRepresentations: []concepts.Concept{{
		UUID:           concordedBrandWithTMEParentOnlyTMEParentUIUD,
		PrefLabel:      "Concept PrefLabel",
		Type:           "Brand",
		Authority:      "TME",
		AuthorityValue: concordedBrandWithTMEParentOnlyTMEParentUIUD,
	}},
}

// Test 6 - Concordance with a TME brand with children but Smartlogic has none so none are surfaced
var concordedBrandWithTMEChildOnlyUUID = "af4e2ccc-a546-41ec-9cb2-db9a4a3f9999"
var concordedBrandWithTMEChildOnlyTMEChildUIUD = "fc16c4b0-a7c8-4454-805f-83a2fc975e5d"
var concordedBrandWithSLChildOnlyUUID = "7e56db9f-ecc1-435a-8c8e-4765fa7cfef9"
var concordedBrandWithTMEChildOnlyAPIOutput = Brand{
	Thing: Thing{mapper.IDURL(concordedBrandWithSLChildOnlyUUID), "http://test.api.ft.com/brands/" + concordedBrandWithSLChildOnlyUUID,
		unfilteredTypes, filterToMostSpecificType(nodeLabels), "Smarto Logico Concorded with one TME with Parents"},
}

var concordedBrandWithTMEChildOnlyConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: concordedBrandWithSLChildOnlyUUID, PrefLabel: "Smarto Logico Concorded with one TME with Parents", Type: "Brand", SourceRepresentations: []concepts.Concept{{
		UUID:           concordedBrandWithSLChildOnlyUUID,
		PrefLabel:      "Smarto Logico Concorded with one TME with Parents",
		Type:           "Brand",
		Authority:      "Smartlogic",
		AuthorityValue: concordedBrandWithSLChildOnlyUUID,
	}, {
		UUID:           concordedBrandWithTMEChildOnlyUUID,
		PrefLabel:      "Concorded with one TME with Parent",
		Type:           "Brand",
		Authority:      "TME",
		AuthorityValue: "99999-XYZ",
	}},
}

var concordedBrandWithTMEChildOnlyParentConceptsWriterInput = concepts.AggregatedConcept{
	PrefUUID: concordedBrandWithTMEChildOnlyTMEChildUIUD, PrefLabel: "Concept PrefLabel", Type: "Brand", SourceRepresentations: []concepts.Concept{{
		UUID:           concordedBrandWithTMEChildOnlyTMEChildUIUD,
		PrefLabel:      "Concept PrefLabel",
		Type:           "Brand",
		Authority:      "TME",
		AuthorityValue: concordedBrandWithTMEChildOnlyTMEChildUIUD,
		ParentUUIDs:    []string{concordedBrandWithTMEChildOnlyUUID},
	}},
}

var oldBrandUUID = "a806e270-edbc-423f-b8db-d21ae90e06c8"

func TestNewConcordanceModelScenarios(t *testing.T) {
	assert := assert.New(t)
	defer cleanDB(t)

	// Setup the model:
	brandsWriter := getConceptsRWDriver(t)

	// Test 1
	_, err := brandsWriter.Write(simpleConceptsWriterInput, "TRANS1")
	assert.NoError(err)

	// Test 2
	_, err = brandsWriter.Write(simpleWithParentParentConceptsWriterInput, "TRANS2")
	assert.NoError(err)
	_, err = brandsWriter.Write(simpleWithParentConceptsWriterInput, "TRANS2")
	assert.NoError(err)

	// Test 3
	_, err = brandsWriter.Write(concordedBrandWithNoParentConceptsWriterInput, "TRANS3")
	assert.NoError(err)

	// Test 4
	_, err = brandsWriter.Write(concordedBrandBothWithParentsSLParentConceptsWriterInput, "TRANS4")
	assert.NoError(err)
	_, err = brandsWriter.Write(concordedBrandBothWithParentsTMEParentConceptsWriterInput, "TRANS4")
	assert.NoError(err)
	_, err = brandsWriter.Write(concordedBrandBothWithParentsConceptsWriterInput, "TRANS4")
	assert.NoError(err)

	// Test 5
	_, err = brandsWriter.Write(concordedBrandWithTMEParentOnlyParentConceptsWriterInput, "TRANS5")
	assert.NoError(err)
	_, err = brandsWriter.Write(concordedBrandWithTMEParentOnlyConceptsWriterInput, "TRANS5")
	assert.NoError(err)

	// Test 6
	_, err = brandsWriter.Write(concordedBrandWithTMEChildOnlyParentConceptsWriterInput, "TRANS6")
	assert.NoError(err)
	_, err = brandsWriter.Write(concordedBrandWithTMEChildOnlyConceptsWriterInput, "TRANS6")
	assert.NoError(err)

	tests := []struct {
		testName      string
		expectedBrand Brand
		brandUUID     string
	}{
		{
			"1. Simple Smartlogic Brand with no parent nor child, representing a new brand added to Smartlogic", simpleAPIOutput, simpleSLBrandUUID,
		},
		{
			"2a. Simple Smartlogic Brand representing a new brand added to Smartlogic with a parent", simpleWithParentAPIOutput, simpleSLBrandWithParentUUID,
		},
		{
			"2b. Simple Smartlogic Brand with a child", simpleWithChildAPIOutput, simpleSLBrandParentsUUID,
		},
		{
			"3. Concordance with a TME brand but neither having a parent", concordedBrandWithNoParentAPIOutput, concordedBrandWithNoParentUUID,
		},
		{
			"4. Concordance with a TME brand both having parents and preferring Smartlogic", concordedBrandBothWithParentsAPIOutput, concordedBrandWithParentsUUID,
		},
		{
			"5. Concordance with a TME brand having a parent but Smartlogic has none so none is returned", concordedBrandWithTMEParentOnlyAPIOutput, concordedBrandWithTMEParentOnlyUUID,
		},
		{
			"6. Concordance with a TME brand having a child but Smartlogic has none so none is returned", concordedBrandWithTMEChildOnlyAPIOutput, concordedBrandWithTMEChildOnlyUUID,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			readAndCompare(t, test.expectedBrand, test.brandUUID)
		})
	}
}

func readAndCompare(t *testing.T, expected Brand, uuid string) {
	srv := getBrandDriver(t)
	srv.env = "test"

	brandFromDB, _, found, err := srv.Read(uuid)

	types := brandFromDB.Types
	assert.NotEmpty(t, brandFromDB)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.NotEmpty(t, brandFromDB)
	assert.Equal(t, expected.Thing.ID, brandFromDB.Thing.ID, fmt.Sprintf("Ids not equal: \n Expected: %v \n Actual: %v", expected.Thing.ID, brandFromDB.Thing.ID))
	assert.Equal(t, expected.Thing.APIURL, brandFromDB.Thing.APIURL, fmt.Sprintf("Api Urls not equal: \n Expected: %v \n Actual: %v", expected.Thing.APIURL, brandFromDB.Thing.APIURL))
	assert.Equal(t, expected.PrefLabel, brandFromDB.PrefLabel, fmt.Sprintf("Pref Label not equal: \n Expected: %v \n Actual: %v", expected.PrefLabel, brandFromDB.PrefLabel))
	assert.Equal(t, expected.DescriptionXML, brandFromDB.DescriptionXML, fmt.Sprintf("Description XML not equal: \n Expected: %v \n Actual: %v", expected.DescriptionXML, brandFromDB.DescriptionXML))
	assert.Equal(t, expected.Strapline, brandFromDB.Strapline, fmt.Sprintf("Strapline not equal: \n Expected: %v \n Actual: %v", expected.Strapline, brandFromDB.Strapline))
	assert.Equal(t, expected.ImageURL, brandFromDB.ImageURL, fmt.Sprintf("Image URLs not equal: \n Expected: %v \n Actual: %v", expected.ImageURL, brandFromDB.ImageURL))
	assert.Equal(t, expected.Types, types, fmt.Sprintf("Types not equal: \n Expected: %v \n Actual: %v", expected.Types, types))
	assert.Equal(t, len(expected.Children), len(brandFromDB.Children))
	for _, expChild := range expected.Children {
		for _, actChild := range brandFromDB.Children {
			if expChild.ID == actChild.ID {
				assert.Equal(t, expChild.ID, actChild.ID, fmt.Sprintf("Child Ids not equal: \n Expected: %v \n Actual: %v", expChild.ID, actChild.ID))
				assert.Equal(t, expChild.PrefLabel, actChild.PrefLabel, fmt.Sprintf("Child Pref Labels not equal: \n Expected: %v \n Actual: %v", expChild.PrefLabel, actChild.PrefLabel))
				assert.Equal(t, expChild.APIURL, actChild.APIURL, fmt.Sprintf("Child Api Urls not equal: \n Expected: %v \n Actual: %v", expChild.APIURL, actChild.APIURL))
			}
		}
	}

	if expected.Parent != nil {
		assert.Equal(t, expected.Parent.ID, brandFromDB.Parent.ID, fmt.Sprintf("Parent Id not equal: \n Expected: %v \n Actual: %v", expected.Parent.ID, brandFromDB.Parent.ID))
		assert.Equal(t, expected.Parent.PrefLabel, brandFromDB.Parent.PrefLabel, fmt.Sprintf("Parent Pref Label not equal: \n Expected: %v \n Actual: %v", expected.Parent.PrefLabel, brandFromDB.Parent.PrefLabel))
		assert.Equal(t, expected.Parent.APIURL, brandFromDB.Parent.APIURL, fmt.Sprintf("Parent Api Url not equal: \n Expected: %v \n Actual: %v", expected.Parent.APIURL, brandFromDB.Parent.APIURL))
	} else {
		assert.Nil(t, brandFromDB.Parent, "No expected Parent yet found a parent")
	}
}

func getConceptsRWDriver(t *testing.T) concepts.ConceptService {
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

func cleanDB(t *testing.T) {
	cleanSourceNodes(t, oldBrandUUID, simpleSLBrandUUID, simpleSLBrandWithParentUUID, simpleSLBrandParentsUUID, concordedBrandWithNoParentUUID, concordedBrandWithNoParentTMEUUID, concordedBrandWithParentsUUID, smartLogicParentUUID, TMEParentUIUD, concordedBrandWithParentsTMEUUID, concordedBrandWithTMEParentOnlyUUID, concordedBrandWithTMEParentOnlyTMEUUID, concordedBrandWithTMEParentOnlyTMEParentUIUD, concordedBrandWithTMEChildOnlyUUID, concordedBrandWithTMEChildOnlyTMEChildUIUD, concordedBrandWithSLChildOnlyUUID)
	deleteSourceNodes(t, oldBrandUUID, simpleSLBrandUUID, simpleSLBrandWithParentUUID, simpleSLBrandParentsUUID, concordedBrandWithNoParentUUID, concordedBrandWithNoParentTMEUUID, concordedBrandWithParentsUUID, smartLogicParentUUID, TMEParentUIUD, concordedBrandWithParentsTMEUUID, concordedBrandWithTMEParentOnlyUUID, concordedBrandWithTMEParentOnlyTMEUUID, concordedBrandWithTMEParentOnlyTMEParentUIUD, concordedBrandWithTMEChildOnlyUUID, concordedBrandWithTMEChildOnlyTMEChildUIUD, concordedBrandWithSLChildOnlyUUID)
	cleanConcordedNodes(t, oldBrandUUID, simpleSLBrandUUID, simpleSLBrandWithParentUUID, simpleSLBrandParentsUUID, concordedBrandWithNoParentUUID, concordedBrandWithNoParentTMEUUID, concordedBrandWithParentsUUID, smartLogicParentUUID, TMEParentUIUD, concordedBrandWithParentsTMEUUID, concordedBrandWithTMEParentOnlyUUID, concordedBrandWithTMEParentOnlyTMEUUID, concordedBrandWithTMEParentOnlyTMEParentUIUD, concordedBrandWithTMEChildOnlyUUID, concordedBrandWithTMEChildOnlyTMEChildUIUD, concordedBrandWithSLChildOnlyUUID)
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
