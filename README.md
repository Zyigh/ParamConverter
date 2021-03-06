# ParamConverter

A simple net/http compatible middleware written in Go 1.16 which aims to facilitate conversion of parameters into usable
object.
It is highly inspired by a [Symfony ParamConverter](https://symfony.com/doc/current/bundles/SensioFrameworkExtraBundle/annotations/converters.html)
usage I have.

## Disclaimer

I am a rookie in Golang, so the implementation is far from perfect. If you see flaws or have any recommendation to
improve the quality of this code, please open an issue and detail as much as you can.

## What is do

When handling a HTTP Request, dealing with parameters can be very painful and repetitive. It can easily lead to a
terrible mess adding a lot of logic inside a *controller*. In a *less worse* scenario, it forces to handle too many
things inside a dedicated service.

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
	nStr := r.URL.Query().Get("n")
	
	if nStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
    }
    
    n, err := strconv.Atoi(nStr)
    
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return	
    }
    
    // Do smth with n FINNALY
}

func main() {
    mux := http.NewServeMux()
    mux.Handle("/", http.HandlerFunc(myHandler))
    
    log.Fatal(http.ListenAndServe("0.0.0.0:80", mux))
}
```

The main purpose of this library is to provide a middleware that will take the query and populate a dedicated facade to
handle these parameters. Every parameter sent in the URL, a urlencoded form, a multipart form or a json request are 
converted in a `map[string]interface{}` that you can use in a `Deserialize` method of a Facade you defined.

You can then handle parameters in a dedicated workspace. If the binding of the Facade should fail, the middleware would 
stop and return a HTTP BadRequest error.

```go
type myFacade struct {
	n int
}

func (m *myFacade) Deserialize(data map[string]interface{}) error {
	raw, ok := data["n"]
	if !ok {
		return fmt.Errorf(`parameter "n" not found in query`)
	}

	n, err := strconv.Atoi(raw.(string))
	if err != nil {
		return fmt.Errorf("cannot parse n as int\n%s", err.Error())
	}

	m.n = n
	return nil
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	facade, ok := r.Context().Value(paramconverter.FacadeCtxKey).(*myFacade)
	if !ok {
		// Here an InternalServerError is returned because it means smth went wrong with the app, not the conversion
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Do smth with facade
}

func main() {
	mux := http.NewServeMux()
	ctrl := http.HandlerFunc(myHandler)
	mux.Handle("/", paramconverter.New(&myFacade{}, ctrl))

	log.Fatal(http.ListenAndServe("0.0.0.0:80", mux))
}
```

This architecture helps you have a clear separation between parameters handling, controller and business logic. Using a 
Facade object with clear types prevents from dealing with the `http.Request` little by little. 

## How to use

The example above shows a basic usage of this middleware. However, using the paramconverter isn't much more complicated.
There are a few things to note

### New(FacadeInterface, http.Handler)

In order to prevent too much memory usage, the middleware will attach the instance of the concrete `FacadeInterface`
passed to the `func New(FacadeInterface, http.Handler) http.Handler`. So it is recommended to pass a pointer of this 
concrete type as a parameter. Note that the method Deserialize will have to be on a pointer of the struct instead of on
the struct directly to implement `FacadeInterface`.

```go
type myFacade struct{}

// OK
func (m *myFacade) Deserialize(data map[string]interface{}) error {
	// Do things, will populate the Facade parameters used by the middleware
}

// NOT OK
func (m myFacade) Deserialize(data map[string]interface{}) error {
	// Do things, but won't populate the Facade parameters used by the middleware
}

func main() {
	middleware := paramconverter.New(&myFacade{}, handler)
}
```

### data map[string]interface{}

The type `map[string]interface{}` of `data` passed to `Deserialize` method is a bit strange but I couldn't find anything
better for one reason. `url.Values` are easy to handle as it is always a `map[string][]string` you are dealing with.
Handling JSON values, confronts you to a type that can't be described better than `map[string]interface{}`.

This allows to handle multiple values for same form key as usually done in HTML. These were arbitrary choices, and it is
**highly encouraged to discuss them in issues** to improve this library.

* Using same name for multiple values

```html
<input name="param" value="1" />
<input name="param" value="2" />
```

Will give you

```go
map[string]interface{}{
	"param": "1"
}
```

* Using html form array to send multiple values

```html
<input name="param[]" value="1" />
<input name="param[]" value="2" />
```

Will give you

```go
map[string]interface{}{
	"param": []{"1", "2"}
}
```

* Using json

```json

{
  "param": 1
}
```

Will give you

```go
map[string]interface{}{
	"param": 1
}
```

It is important to know that url query is the first to be parsed, so it will be orverridden by values with same name in 
form or json.

```html
<form action="/?param=1">
    <input name="param" value="8" />
</form>
```

Will give you

```go
map[string]interface{}{
	"param": 8
}
```

### Parameters

Some parameters can be defined by the user

* DefaultMultipartMaxMemory int64

The value pass to `http.Request.ParseMultipartForm(maxMemory int64) error`. The default value is `0` because this
middleware doesn't deal with files, but you can change it globally in your code.

```go
func main() {
	paramconverter.DefaultMultipartMaxMemory = 42000000
}
```

An improvement of the library could be to handle this parameter by request but it seems as first intuition that it will
make the code look bigger. Don't hesitate to recommend or propose implementation in issues.

* FacadeCtxKey interface{}

Is the key with which you get the Facade from `http.Request.Context()`. It is by default a struct defined in
paramConverter.go (`facadeCtxKey`), but feel free to change it if you need. 

## Contribute

Feel free to report bugs and suggest improvements in issues. There isn't any PR template as it doesn't seem necessary
yet, but try to make it clear enough as you would make for any other project. PR won't be accepted if not related to an
issue though.
