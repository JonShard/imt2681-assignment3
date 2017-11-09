package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/Arxcis/imt2681-assignment2/lib/database"
	"github.com/Arxcis/imt2681-assignment2/lib/types"
	"gopkg.in/mgo.v2/bson"
)

var APP *App

func init() {

	configpath := "../../../config/currency.json"
	APP = &App{
		Path:              "/api/test",
		Port:              "5555",
		CollectionWebhook: "testhook",
		CollectionFixer:   "testfixer",
		Mongo: database.Mongo{
			Name:    "test",
			URI:     "127.0.0.1:33017",
			Session: nil,
		},
		Currency: func() []string {
			log.Println("Reading " + configpath)
			data, err := ioutil.ReadFile(configpath)
			if err != nil {
				panic(err.Error())
			}
			var currency []string
			if err = json.Unmarshal(data, &currency); err != nil {
				panic(err.Error())
			}
			log.Println("Done with " + configpath)
			return currency
		}(),
	}
	// @verbose
	// indented, _ := json.MarshalIndent(APP, "", "    ")
	// log.Println(string(indented))
	log.Println("Webhookserver initialized...")

	log.Println("Reseeding DB")
	reseedDB()
	log.Println("Done with DB")
}

func TestHelloWorld(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(APP.HelloWorld))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err.Error())
	}

	fmt.Printf("%s", greeting)
}

var testid bson.ObjectId

func TestPostWebhook(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(APP.PostWebhook))
	defer ts.Close()
	json, err := json.Marshal(types.Webhook{
		WebhookURL:      ts.URL,
		BaseCurrency:    "EUR",
		TargetCurrency:  "NOK",
		MinTriggerValue: 7.7,
		MaxTriggerValue: 9.9,
	})

	res, err := http.Post(ts.URL, "application/json", bytes.NewReader(json))
	if err != nil {
		t.Error(err.Error())
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err.Error())
	}

	fmt.Printf("%s\n", greeting)
}
func TestGetWebhook(t *testing.T) {

	r := mux.NewRouter()
	r.HandleFunc(APP.Path+"/webhook/{id}", APP.GetWebhook).Methods("GET")
	ts := httptest.NewServer(r)
	defer ts.Close()
	ids := map[string]int{
		//"":           http.StatusBadRequest,
		bson.NewObjectId().Hex(): http.StatusNotFound,
		testid.Hex():             http.StatusOK,
	}

	for id, status := range ids {
		url := ts.URL + APP.Path + "/webhook/" + id
		log.Println("url: ", url)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatal(err)
		}
		if s := resp.StatusCode; s != status {
			t.Fatalf("Wrong status code. Got %d want %d", s, status)
		} else {
			data, _ := ioutil.ReadAll(resp.Body)
			log.Println(string(data))
		}
	}

}
func TestGetWebhookAll(t *testing.T)      {}
func TestGetLatestCurrency(t *testing.T)  {}
func TestGetAverageCurrency(t *testing.T) {}
func TestEvaluationTrigger(t *testing.T)  {}

func reseedDB() error {

	db, err := APP.Mongo.Open()
	if err != nil {
		return err
	}
	defer APP.Mongo.Close()

	cWebhook := db.C(APP.CollectionWebhook)
	cFixer := db.C(APP.CollectionFixer)

	cFixer.DropCollection()
	cWebhook.DropCollection()

	cFixer.Insert(bson.M{
		"base": "EUR",
		"date": "2017-10-24",
		"rates": map[string]float64{
			"NOK": 9.3883,
			"TRY": 4.3751,
			"USD": 1.1761,
			"ZAR": 16.14,
		},
	}, bson.M{
		"base": "EUR",
		"date": "2017-10-23",
		"rates": map[string]float64{
			"NOK": 9.3883,
			"TRY": 4.3751,
			"USD": 1.1761,
			"ZAR": 16.14,
		},
	}, bson.M{
		"base": "EUR",
		"date": "2017-10-22",
		"rates": map[string]float64{
			"NOK": 9.3883,
			"TRY": 4.3751,
			"USD": 1.1761,
			"ZAR": 16.14,
		},
	})
	testid = bson.NewObjectId()
	cWebhook.Insert(bson.M{
		"_id":             testid,
		"webhookURL":      "127.0.0.1:5555",
		"baseCurrency":    "EUR",
		"targetCurrency":  "NOK",
		"minTriggerValue": 9.0,
		"maxTriggerValue": 9.9,
	}, bson.M{
		"webhookURL":      "127.0.0.1:5555",
		"baseCurrency":    "EUR",
		"targetCurrency":  "NOK",
		"minTriggerValue": 9.0,
		"maxTriggerValue": 9.9,
	})
	return nil
}
