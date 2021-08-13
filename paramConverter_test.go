package paramconverter_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"paramconverter"
	"strconv"
	"strings"
	"testing"
)

/// Setup

type facadeTest struct {
	param int
}

func (f *facadeTest) Deserialize(data map[string]interface{}) error {
	raw, ok := data["param"]

	if !ok {
		return fmt.Errorf(`parameter "param" not found in query`)
	}

	param, err := strconv.Atoi(raw.(string))

	if err != nil {
		return fmt.Errorf("cannot parse param as int\n%s", err.Error())
	}

	f.param = param

	return nil
}

type emptyFacade struct {}

func (e *emptyFacade) Deserialize(map[string]interface{}) error {
	return nil
}

func httpTestHandler(w http.ResponseWriter, r *http.Request) {
	facade, ok := r.Context().Value(paramconverter.FacadeCtxKey).(*facadeTest)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d",facade.param)
}

func httpEmptyTestHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(paramconverter.FacadeCtxKey).(*emptyFacade)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

/// Tests

func TestParamConverterConvertsGetRequest(t *testing.T) {
	expected := "1"

	req, err := http.NewRequest("GET", "/?param=" + expected, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := paramconverter.New(&facadeTest{}, http.HandlerFunc(httpTestHandler))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HTTP request failed, expected status 200, got status %d", status)
	}

	if recorder.Body.String() != expected {
		t.Errorf("bad body returned, expected %s, got %s", expected, recorder.Body.String())
	}
}

func TestParamConverterReturnsBadRequestWhenFailsToConvertGetRequest(t *testing.T) {
	str := "not+an+int"
	req, err := http.NewRequest("GET", "/?param="+str, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := paramconverter.New(&facadeTest{}, http.HandlerFunc(httpTestHandler))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("HTTP request failed, expected status 400, got status %d", status)
	}

	if len(recorder.Body.String()) != 0 {
		t.Errorf("unexpected body, expected empty, got %s", recorder.Body.String())
	}
}

func TestParamConverterConvertsJsonPostRequest(t *testing.T) {
	data := map[string]string{
		"param": "1",
	}
	expected := "1"

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")

	handler := paramconverter.New(&facadeTest{}, http.HandlerFunc(httpTestHandler))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HTTP request failed, expected status 200, got status %d", status)
	}

	if recorder.Body.String() != expected {
		t.Errorf("bad body returned, expected %s, got %s", expected, recorder.Body.String())
	}
}

func TestParamConverterHandlesEmptyJson(t *testing.T)  {
	expected := ""

	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")

	handler := paramconverter.New(&emptyFacade{}, http.HandlerFunc(httpEmptyTestHandler))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HTTP request failed, expected status 200, got status %d", status)
	}

	if recorder.Body.String() != expected {
		t.Errorf("bad body returned, expected %s, got %s", expected, recorder.Body.String())
	}
}

func TestParamConverterConvertsPostHtmlForms(t *testing.T) {
	expected := "1"

	form := url.Values{}
	form.Set("param", expected)

	encoded := form.Encode()

	req, err := http.NewRequest("POST", "/", strings.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(encoded)))

	handler := paramconverter.New(&facadeTest{}, http.HandlerFunc(httpTestHandler))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HTTP request failed, expected status 200, got status %d", status)
	}

	if recorder.Body.String() != expected {
		t.Errorf("bad body returned, expected %s, got %s", expected, recorder.Body.String())
	}
}

func TestParamConverterMiddlewareDoesntFailOnEmptyForm(t *testing.T) {
	expected := ""

	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "0")

	handler := paramconverter.New(&emptyFacade{}, http.HandlerFunc(httpEmptyTestHandler))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HTTP request failed, expected status 200, got status %d", status)
	}

	if recorder.Body.String() != expected {
		t.Errorf("bad body returned, expected %s, got %s", expected, recorder.Body.String())
	}
}
