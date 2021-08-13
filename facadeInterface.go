package paramconverter

// FacadeInterface interface that must be implemented by structs to bind with query parsed with the param converter
type FacadeInterface interface {
	// Deserialize method that takes data (all url query, form, multipart form and json request) keys and values to be
	// converted into parameters of a FacadeInterface. Note that data has to be a map[string]interface as it is the less
	// specific typing of a raw json.
	//
	// The main point is to transform the query you expect into an instance of a something you defined, and deal with
	// bad queries (such as wrong type, invalid data...) before these data are handled in the controller
	//
	// If an error is returned, it will be loged, and the web application will just return a HTTP error 400
	Deserialize(data map[string]interface{}) error
}
