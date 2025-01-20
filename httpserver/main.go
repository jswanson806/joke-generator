package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
		if err != nil {
			fmt.Println("Error getting name:", err)
			return
		}
	}()

	wg.Wait()
	if err != nil {
		//Handle name retrieval error
		http.Error(w, "Failed to get name", http.StatusInternalServerError)
		return
	}

	// Add to WaitGroup
	wg.Add(1)
	// goroutine to get random joke
	//	Pass first and last name returned from getRandomName()
	go func() {
		defer wg.Done()
		joke, err = getRandomJoke(name.FirstName, name.LastName)
		if err != nil {
			fmt.Println("Error getting joke:", err)
			return
		}
	}()

	wg.Wait()
	if err != nil {
		// Handle joke retrieval error
		http.Error(w, "Failed to get joke", http.StatusInternalServerError)
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
func getRandomName() (Names, error) {
	// Parse randNameEndpoint into a URL structure
	base, err := url.Parse(randNameEndpoint)
	// Handle errors while parsing and exit program
	if err != nil {
		fmt.Printf("client could not parse url: %s\n", err)
		os.Exit(1)
	}
	// Create the GET request
	req, err := http.NewRequest(http.MethodGet, base.String(), nil)
	// Handle errors creating request and exit program
	if err != nil {
		fmt.Printf("client could not create request: %s\n", err)
		os.Exit(1)
	}
	// Timeout if request takes longer than 30 seconds
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	// Make the request
	res, err := client.Do(req)
	// Handle errors while making request and exit program
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}
	// Print client message and status code for debugging
	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	// Read the response body
	resBody, err := io.ReadAll(res.Body)
	// Handle errors while reading response body and exit program
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	// Unmarshal JSON in resBody and initialize struct Names with data
	var n Names
	// Handle errors while unmarshalling resBody JSON and exit program
	if err := json.Unmarshal(resBody, &n); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		os.Exit(1)
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
func getRandomJoke(firstName, lastName string) (string, error) {
	// Parse randJokeBaseEndpoint into a URL structure
	base, err := url.Parse(randJokeBaseEndpoint)
	// Handle errors while parsing url and exit program
	if err != nil {
		fmt.Printf("client could not parse url: %s\n", err)
		os.Exit(1)
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
		fmt.Printf("client could not create request: %s\n", err)
		os.Exit(1)
	}
	// Timeout if request takes longer than 30 seconds
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	// Make the request
	res, err := client.Do(req)
	// Handle errors while making request and exit program
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}
	// Print client message and status code for debugging
	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	// Read the response body
	resBody, err := io.ReadAll(res.Body)
	// Handle errors while reading response body and exit program
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	// Initialize new Joke struct
	var j Joke
	// Unmarshal JSON in resBody and initialize struct Names with data
	if err := json.Unmarshal(resBody, &j); err != nil {
		// Handle errors while unmarshalling resBody JSON and exit program
		fmt.Println("Error unmarshalling JSON:", err)
		os.Exit(1)
	}
	// Return joke string from Joke struct
	return j.Value.Joke, nil
}

/*
	 Function writes the joke string to http.ResponseWriter

		Accepts a string and Writer
*/
func returnCompleteJoke(joke string, w http.ResponseWriter) {
	// Write joke string
	_, err := io.WriteString(w, joke)
	// Handle errors while writing response
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
