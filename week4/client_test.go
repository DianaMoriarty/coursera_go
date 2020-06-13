package main

import (
	"encoding/json"
	"encoding/xml"
	_ "fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

type Users struct {
	Id        int    `xml:"id"`
	Guid      string `xml:"guid"`
	IsActive  bool   `xml:"isActive"`
	Balance   string `xml:"balance"`
	Picture   string `xml:"picture"`
	Age       int    `xml:"age"`
	EyeColor  string `xml:"eyeColor"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
	Company   string `xml:"company"`
	Email     string `xml:"email"`
	Phone     string `xml:"phone"`
	Address   string `xml:"address"`
	About     string `xml:"about"`
}

type xmlStruct struct {
	Version string  `xml:"version"`
	Row     []Users `xml:"row"`
}

func SearchServerSuccess(w http.ResponseWriter, r *http.Request) {
	var users []User
	var s int
	file, err := ioutil.ReadFile("dataset.xml")
	checkError(err)
	usersXml := &xmlStruct{}
	xml.Unmarshal(file, &usersXml)
	for _, user := range usersXml.Row {
		users = append(users, User{
			Id:     user.Id,
			Name:   user.FirstName,
			Age:    user.Age,
			About:  user.About,
			Gender: user.Gender,
		})
	}
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if offset > 0 {
		s = offset * 25
	}
	users = users[s:s + limit]
	jsonResp, err := json.Marshal(users)
	checkError(err)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func SearchServerLimitFail(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.ReadFile("dataset.xml")
	checkError(err)
	usersXml := &xmlStruct{}
	xml.Unmarshal(file, &usersXml)
	var users []User
	for _, user := range usersXml.Row {
		users = append(users, User{
			Id:     user.Id,
			Name:   user.FirstName,
			Age:    user.Age,
			About:  user.About,
			Gender: user.Gender,
		})
	}
	jsonResponse, err := json.Marshal(users)
	checkError(err)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func SearchServerJsonFail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `"err": "Json fail"}`)
}

func SearchServerTimeoutError(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 2)
	w.WriteHeader(http.StatusOK)
}

func SearchServerUnknownError(w http.ResponseWriter, r *http.Request) {}

func SearchServerUnauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}

func SearchServerInternalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func SearchServerBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func SearchServerBadField(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	resp, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
	w.Write(resp)
}

func SearchServerBadError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonResponse, _ := json.Marshal(SearchErrorResponse{Error: "Unknown error"})
	w.Write(jsonResponse)
}

func TestErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerSuccess))
	searchClient := &SearchClient{URL: ts.URL,}
	searchRequest := SearchRequest{Limit:  5, Offset: 0,}
	_, err := searchClient.FindUsers(searchRequest)
	if err != nil {
		t.Error("Dosn't work success request")
	}
	searchRequest.Limit = -1
	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "limit must be > 0" {
		t.Error("limit must be > 0")
	}
	searchRequest = SearchRequest{Limit:  1, Offset: -1,}
	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "offset must be > 0" {
		t.Error("offset must be > 0")
	}
	ts.Close()
}

func TestLimitFailed(t *testing.T) {
	limit := 7
	ts := httptest.NewServer(http.HandlerFunc(SearchServerLimitFail))
	searchClient := &SearchClient{URL: ts.URL}
	if resp, _ := searchClient.FindUsers(SearchRequest{Limit: limit}); limit == len(resp.Users) {
		t.Error("Limit not true")
	}
	ts.Close()
}

func TestBadJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerJsonFail))
	searchClient := &SearchClient{URL: ts.URL,}
	if _, err := searchClient.FindUsers(SearchRequest{}); err.Error() != `cant unpack result json: invalid character ':' after top-level value` {
		t.Error("Bad json test :(")
	}
	ts.Close()
}

func TestPerelimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerSuccess))
	searchClient := &SearchClient{URL: ts.URL,}
	if resp, _ := searchClient.FindUsers(SearchRequest{Limit: 26});25 != len(resp.Users) {
		t.Error("Perelimit :(")
	}
	ts.Close()
}

func TestTimeoutError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerTimeoutError))
	searchClient := &SearchClient{URL: ts.URL,}
	if _, err := searchClient.FindUsers(SearchRequest{}); err == nil {
		t.Error("Timeout chck error :(")
	}
	ts.Close()
}

func TestUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerUnknownError))
	searchClient := &SearchClient{URL: "bad_link",}
	if _, err := searchClient.FindUsers(SearchRequest{}); err == nil {
		t.Error("TestUnknownError :(")
	}
	ts.Close()
}

func TestStatusUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerUnauthorized))
	searchClient := &SearchClient{URL: ts.URL}
	if _, err := searchClient.FindUsers(SearchRequest{}); err.Error() != "Bad AccessToken" {
		t.Error("Bad AccessToken is not done :(")
	}
	ts.Close()
}

func TestStatusInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerInternalServerError))
	searchClient := &SearchClient{URL: ts.URL}
	if _, err := searchClient.FindUsers(SearchRequest{}); err.Error() != "SearchServer fatal error" {
		t.Error("SearchServer fatal error is not done :(")
	}
	ts.Close()
}

func TestBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadRequest))
	searchClient := &SearchClient{URL: ts.URL}
	if _, err := searchClient.FindUsers(SearchRequest{}); err.Error() != "cant unpack error json: unexpected end of JSON input" {
		t.Error("TestBadRequest is not done")
	}
	ts.Close()
}

func TestBadField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadField))
	searchClient := &SearchClient{URL: ts.URL}
	if _, err := searchClient.FindUsers(SearchRequest{}); err.Error() != "OrderFeld  invalid" {
		t.Error("ErrorBadOrderField is not done")
	}
	ts.Close()
}

func TestBadRequestError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadError))
	searchClient := &SearchClient{URL: ts.URL}
	if _, err := searchClient.FindUsers(SearchRequest{}); err == nil {
		t.Error("TestBadRequestError is not done")
	}
	ts.Close()
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
