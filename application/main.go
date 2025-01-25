package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const serverPort = 3000

// Endpoint for getting a random first and last name
const randNameEndpoint = "https://names.mcquay.me/api/v0/"

// Base endpoint for generating a random joke.
// Use query string values 'firstName' and 'lastName' to personalize
const randJokeBaseEndpoint = "http://joke.loc8u.com:8888/joke?limitTo=nerdy"

// struct to hold expected output of Names
type Names struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// struct to hold expected output of Joke
type Joke struct {
	Value struct {
		Joke string `json:"joke"`
	} `json:"value"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var name Names
	var joke string
	var err error

	// Add to WaitGroup
	wg.Add(1)
	// goroutine to get random first and last name
	go func() {
		defer wg.Done()
		name, err = getRandomName()
		// Set error while getting name
		if err != nil {
			err = fmt.Errorf("failed to get name: %w", err)
			return
		}
	}()

	wg.Wait()
	//Handle name retrieval error
	if err != nil {
		http.Error(w, "failed to get name", http.StatusInternalServerError)
		return
	}

	// Add to WaitGroup
	wg.Add(1)

	// goroutine to get random joke
	//	Pass first and last name returned from getRandomName()
	go func() {
		defer wg.Done()
		joke, err = getRandomJoke(name.FirstName, name.LastName)
		// Handle error while getting joke
		if err != nil {
			err = fmt.Errorf("error getting joke: %w", err)
			return
		}
	}()

	wg.Wait()
	// Handle joke retrieval error
	if err != nil {
		http.Error(w, "failed to get joke", http.StatusInternalServerError)
		return
	}

	// Call function to return completed joke
	returnCompleteJoke(joke, w)
}

func main() {

	// Use http.ServeMux struct instead of default multiplexer
	mux := http.NewServeMux()

	// Handlers for routes are defined below
	mux.HandleFunc("/", getRoot)

	// Set up the server
	server := http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", serverPort),
		Handler: mux,
	}

	// Start server with parameters configured above for server
	err := server.ListenAndServe()

	// Handle ErrServerClosed error
	if !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error running http server: %s\n", err)
	}
}

/*
	 Function to return random first and last name.

		Calls external web service:
			"https://names.mcquay.me/api/v0/"

		Returns Names struct
*/
var getRandomName = func() (Names, error) {
	// Parse randNameEndpoint into a URL structure
	base, err := url.Parse(randNameEndpoint)
	// Handle errors while parsing
	if err != nil {
		return Names{}, fmt.Errorf("client could not parse url: %s", err)
	}
	// Create the GET request
	req, err := http.NewRequest(http.MethodGet, base.String(), nil)
	// Handle errors creating request
	if err != nil {
		return Names{}, fmt.Errorf("client could not create request: %s", err)
	}
	// Timeout if request takes longer than 30 seconds
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	// Make the request
	res, err := client.Do(req)
	// Handle errors while making request
	if err != nil {
		return Names{}, fmt.Errorf("client: error making http request: %s", err)
	}
	// Print client message and status code for debugging
	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	// Read the response body
	resBody, err := io.ReadAll(res.Body)
	// Handle errors while reading response body
	if err != nil {
		return Names{}, fmt.Errorf("client: could not read response body: %s", err)
	}
	// Initialize struct to hold return values
	var n Names
	// Verify response body is valid JSON
	if json.Valid(resBody) {
		// Unmarshal JSON in resBody and initialize struct Names with data

		// Handle errors while unmarshalling resBody JSON and exit program
		if err := json.Unmarshal(resBody, &n); err != nil {
			return Names{}, fmt.Errorf("error unmarshalling JSON: %s", err)
		}
		// If not valid JSON, handle error and print body
		//	does not cause failure state
	} else {
		return Names{}, fmt.Errorf("non-JSON response received: %s", string(resBody))
	}
	// Return Names struct
	return n, err
}

/*
	 Function to return random Chuck Norris joke

		Accepts firstName and lastName as arguments
		and calls external web service:
			"http://joke.loc8u.com:8888/joke?limitTo=nerdy"

		Passes firstName and lastName in the query string to
		personalize the joke being returned.

		Returns Joke struct
*/
var getRandomJoke = func(firstName, lastName string) (string, error) {
	// Parse randJokeBaseEndpoint into a URL structure
	base, err := url.Parse(randJokeBaseEndpoint)

	// Handle errors while parsing url
	if err != nil {
		return "", fmt.Errorf("client could not parse url: %s", err)
	}

	// Initialize Values map 'params'
	params := url.Values{}

	// Add the firstName and lastName to params
	params.Add("firstName", firstName)
	params.Add("lastName", lastName)

	// Encode and add query string values to base URL
	base.RawQuery = params.Encode()

	// Create the GET request
	req, err := http.NewRequest(http.MethodGet, base.String(), nil)

	// Handle errors while creating the request and exit program
	if err != nil {
		return "", fmt.Errorf("client could not create request: %s", err)
	}

	// Timeout if request takes longer than 30 seconds
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	res, err := client.Do(req)

	// Handle errors while making request and exit program
	if err != nil {
		return "", fmt.Errorf("client: error making http request: %s", err)
	}

	// Print client message and status code for debugging
	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	// Read the response body
	resBody, err := io.ReadAll(res.Body)

	// Handle errors while reading response body and exit program
	if err != nil {
		return "", fmt.Errorf("client: could not read response body: %s", err)
	}

	// Initialize new Joke struct
	var j Joke

	// Unmarshal JSON in resBody and initialize struct Names with data
	if err := json.Unmarshal(resBody, &j); err != nil {
		// Handle errors while unmarshalling resBody JSON and exit program
		return "", fmt.Errorf("error unmarshalling JSON: %s", err)
	}

	// Return joke string from Joke struct
	return j.Value.Joke, nil
}

/*
	 Function writes the joke string to http.ResponseWriter

		Accepts a string and Writer
*/
var returnCompleteJoke = func(joke string, w http.ResponseWriter) {
	// Write joke string
	_, err := io.WriteString(w, joke)

	// Handle errors while writing response
	if err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
