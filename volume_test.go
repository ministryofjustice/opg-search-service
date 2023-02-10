package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/objx"
	"github.com/stretchr/testify/suite"
)

//This suite is best run with verbose to see the logs of the results/rankings
type HighDataSuite struct {
	suite.Suite
}

// load in data
func (suite *HighDataSuite) SetupSuite() {
	suite.T().Log("Setting up data")
	os.Setenv("ENVIRONMENT", "local")
	//easier to run in both docker and locally
	os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", "http://localstack:4571")
	if runtime.GOOS == "darwin" {
		os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", "http://localhost:9200")
	}

	esClient := getClient()
	ctx := context.Background()
	loadedRecords := 1000

	suite.T().Log("Building index")

	indexExists, _ := esClient.IndexExists(ctx, "person")
	if !indexExists {
		_, config, _ := person.IndexConfig()

		esClient.CreateIndex(ctx, "person", config, true)
	}

	bulk := elasticsearch.NewBulkOp("person")
	for i := 0; i < loadedRecords; i++ {
		ID := int64(i)
		bulk.Index(ID, makePersonRecord(ID))
	}
	suite.T().Logf("Loaded %d records", loadedRecords)

	esClient.DoBulk(ctx, bulk)

}

func getClient() *elasticsearch.Client {
	logger := logrus.New()
	httpClient := &http.Client{}
	esClient, _ := elasticsearch.NewClient(httpClient, logger)
	return esClient
}

func (suite *HighDataSuite) TestCommonName() {
	getQueryResult(suite, "John")
}

func (suite *HighDataSuite) TestCommonNamesCombo() {
	getQueryResult(suite, "John Smith")
}

func (suite *HighDataSuite) TestCommonNamesLocationCombo() {
	getQueryResult(suite, "John Smith London")
}

func (suite *HighDataSuite) TestNameLocationCollisions() {
	getQueryResult(suite, "Jack London")

}

func (suite *HighDataSuite) TestAmbiguousPlaceSurname() {
	getQueryResult(suite, "Jo Green")
}

func (suite *HighDataSuite) TestNameStreetCollisions() {
	getQueryResult(suite, "Jack Street")
}

func (suite *HighDataSuite) TestNameCollisions() {

	getQueryResult(suite, "Jack O'Reilly")

	getQueryResult(suite, "Jack Reilly")
}

func (suite *HighDataSuite) TestIDs() {
	getQueryResult(suite, "7000-0002-0390")
	getQueryResult(suite, "700000020390")
}

func getQueryResult(suite *HighDataSuite, search string) {
	esClient := getClient()
	ctx := context.Background()
	indexes := []string{"person"}
	suite.T().Logf("Testing search for %s", search)

	query := []byte(getQuery(search, false))
	var data map[string]interface{}

	json.Unmarshal(query, &data)

	result, _ := esClient.Search(ctx, indexes, data)
	suite.T().Log("- Unboosted:")
	logResults(suite, result)

	query2 := []byte(getQuery(search, true))
	var data2 map[string]interface{}

	json.Unmarshal(query2, &data2)

	result2, _ := esClient.Search(ctx, indexes, data2)
	suite.T().Log("- Boosted:")
	logResults(suite, result2)
}

func logResults(suite *HighDataSuite, result *elasticsearch.SearchResult) {
	suite.T().Logf("%d results", result.Total)
	for count, item := range result.Hits {

		ob, _ := objx.FromJSON(string(item))
		suite.T().Logf("- %d %s %s %s %s,%s,%s", count+1, ob.Get("firstname"), ob.Get("surname"), ob.Get("uId"), ob.Get("addresses[0].addressLines[0]"), ob.Get("addresses[0].addressLines[1]"), ob.Get("addresses[0].addressLines[2]"))
	}
}

func getQuery(search string, boost bool) string {
	if boost {
		return fmt.Sprintf(`{"query": {"simple_query_string" : {"query":"%s", "fields":["full_name^3", "*"]}}}`, search)
	}
	return fmt.Sprintf(`{"query": {"simple_query_string" : {"query":"%s", "fields":["*"]}}}`, search)
}

func TestHighVolume(t *testing.T) {

	suite.Run(t, new(HighDataSuite))
}

func makePersonRecord(index int64) person.Person {

	//create a person to index
	//built from common collisions lists
	firstNames := [...]string{"John", "Jack", "Jan", "Ian", "Joe", "Jo", "Joseph", "Joanne", "Lois", "Elton", "Jackie", "Johnnie", "Johnny", "Jon", "Jonathan", "Bill", "William"}
	surNames := [...]string{"Lane", "London", "Bedford", "Elton", "John", "Smith", "Street", "O'Reilly", "Reilly", "Derby", "Green"}
	postcodes := [...]string{"NG1 2CD", "NW1", "TL5 5BH", "B31 3LP"}
	streets := [...]string{"Watling Lane", "Smith Street", "Bedford Street", "Derby Road", "London Road", "Green Lane", "The Green", "Bristol Road", "Jack Street", "John Street", "Williams Road", "Elton Road"}
	towns := [...]string{"London", "Bedford", "Derby", "Cornwall", "Bristol", "Elton"}

	randomFirstNameIndex := rand.Intn(len(firstNames))
	randomSurNameIndex := rand.Intn(len(surNames))
	randomPostcode := rand.Intn(len(postcodes))
	randomStreet := rand.Intn(len(streets))
	randomTown := rand.Intn(len(towns))

	addressLines := []string{"10"}

	addressLines = append(addressLines, streets[randomStreet])
	addressLines = append(addressLines, towns[randomTown])

	uid := fmt.Sprintf("70000002%04d", index)
	normalisedUid, _ := strconv.ParseInt(uid, 10, 64)

	return person.Person{
		ID:            &index,
		Firstname:     firstNames[randomFirstNameIndex],
		Surname:       surNames[randomSurNameIndex],
		Persontype:    "Deputy",
		Dob:           fmt.Sprintf("%d/%d/19%d", rand.Intn(28), rand.Intn(12), rand.Intn(99)),
		UID:           fmt.Sprintf("7000-0002-%04d", index),
		Normalizeduid: normalisedUid,
		Addresses: []person.PersonAddress{{
			Addresslines: addressLines,
			Postcode:     postcodes[randomPostcode],
		}},
	}
}
