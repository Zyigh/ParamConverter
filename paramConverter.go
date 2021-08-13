package paramconverter

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// facadeCtxKey default key to store facade in http.Request Context
type facadeCtxKey struct {}

// paramConverter struct that holds the FacadeInterface to use it inside the handler function
type paramConverter struct {
	facade FacadeInterface
}

var (
	// DefaultMultipartMaxMemory Default value to pass to http.Request.ParseMultipartForm(maxMemory int64)
	DefaultMultipartMaxMemory int64 = 0
	// FacadeCtxKey the key that will be used to store the concrete implementation of FacadeInterface in
	// http.Request.Context
	FacadeCtxKey interface{}		= facadeCtxKey{}
)

// extractDataFrom helper to transform url.Values into a map[string]interface compatible with the type of data
// to inject into FacadeInterface.Deserialize. The point of it is to handle array form values defined in HTML like
//
// <input name="input[]" />
// <input name="input[]" />
func (p paramConverter) extractDataFrom(values url.Values, data map[string]interface{}) map[string]interface{} {
	for key, val := range values {
		if strings.HasSuffix(key, "[]") {
			k := strings.TrimSuffix(key, "[]")
			data[k] = val
		} else {
			data[key] = values.Get(key)
		}
	}

	return data
}

// addFacadeToRequest the http middleware in itself. It extracts data from http.Request.URL.Query, then check on the
// http.Request Header "Content-Type" and extracts the data of the json, urlencoded form or multipart form received.
//
// The data extracted from queries are then bound to the FacadeInterface stored on the instance. If an error occurs in
// the process (decoding JSON or Deserializing the FacadeInterface), a http.StatusBadRequest is returned and the next
// middleware will not be called.
func (p paramConverter) addFacadeToRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		data := p.extractDataFrom(r.URL.Query(), map[string]interface{}{})

		switch r.Header.Get("Content-Type") {
		case "application/json":
			if nil != r.Body {
				decoder := json.NewDecoder(r.Body)
				err := decoder.Decode(&data)

				if err != nil {
					log.Printf("Undecodable json\n%s\n", err.Error())
					w.WriteHeader(http.StatusBadRequest)

					return
				}
			}
		case "multipart/form-data":
			if nil == r.ParseMultipartForm(DefaultMultipartMaxMemory) {
				data = p.extractDataFrom(r.MultipartForm.Value, data)
			}
		case "application/x-www-form-urlencoded":
			if nil == r.ParseForm() {
				data = p.extractDataFrom(r.Form, data)
			}
		}

		if err := p.facade.Deserialize(data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Param conversion error: %s\n", err.Error())

			return
		}

		ctx := context.WithValue(r.Context(), FacadeCtxKey, p.facade)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// New instantiate a paramConverter with the facade and returns it's addFacadeToRequest (middleware in itself) method so
// it can be added in the middlewares list.
//
// It can be called like any other middleware, except the first argument is a concrete instance of a FacadeInterface
// e.g.
// handler := paramconverter.New(&ConcreteFacade{}, http.HandlerFunc(helloWorld))
func New(facade FacadeInterface, next http.Handler) http.Handler {
	return paramConverter{facade: facade}.addFacadeToRequest(next)
}
