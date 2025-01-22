package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestGetRoot(t *testing.T) {
	// Save and restore original function implementations
	originalGetRandomName := getRandomName
	originalGetRandomJoke := getRandomJoke
	defer func() {
		getRandomName = originalGetRandomName
		getRandomJoke = originalGetRandomJoke
	}()
	// Mock getRandomName to return a predefined value
	getRandomName = func() (Names, error) {
		return Names{FirstName: "John", LastName: "Doe"}, nil
	}
	// Mock getRandomJoke to return predefined value
	getRandomJoke = func(firstName, lastName string) (string, error) {
		return "Mocked joke about John Doe", nil
	}

	t.Run("Returns 200 status code", func(t *testing.T) {

		// Create a request to pass to the handler
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		// Handle error while creating response
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}

		// Record response
		rec := httptest.NewRecorder()

		// Initialize handler
		handler := http.HandlerFunc(getRoot)

		// Call the handler
		handler.ServeHTTP(rec, req)

		// Check the status code for 200
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK; got %v", rec.Code)
		}
	})

	t.Run("Returns expected string", func(t *testing.T) {

		// Create a request to pass to the handler
		req, err := http.NewRequest(http.MethodGet, "/", nil)

		// Handle error while creating the request
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}

		// Record response
		rec := httptest.NewRecorder()

		// Initialize handler
		handler := http.HandlerFunc(getRoot)

		// Call the handler
		handler.ServeHTTP(rec, req)

		// String expected from the response
		expected := "Mocked joke about John Doe"

		// Verify expected string is in the response body
		if !strings.Contains(rec.Body.String(), expected) {
			t.Errorf("Expected response body to contain %q; got %q", expected, rec.Body.String())
		}
	})
}

func TestGetRootFailures(t *testing.T) {
	// Save and restore original function implementations
	originalGetRandomName := getRandomName
	originalGetRandomJoke := getRandomJoke
	defer func() {
		getRandomName = originalGetRandomName
		getRandomJoke = originalGetRandomJoke
	}()

	t.Run("getRandomName failure", func(t *testing.T) {
		// Mock getRandomName to return an error
		getRandomName = func() (Names, error) {
			return Names{}, fmt.Errorf("failed to fetch name")
		}

		// Create a request to pass to the handler
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}

		// Record response
		rec := httptest.NewRecorder()

		// Initialize the handler
		handler := http.HandlerFunc(getRoot)

		// Call the handler
		handler.ServeHTTP(rec, req)

		// Verify status code is 500
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status Internal Server Error; got %v", rec.Code)
		}
	})

	t.Run("getRandomJoke failure", func(t *testing.T) {
		// Mock getRandomName
		getRandomName = func() (Names, error) {
			return Names{FirstName: "John", LastName: "Doe"}, nil
		}

		// Mock and simulate a failed call to getRandomJoke
		getRandomJoke = func(firstName, lastName string) (string, error) {
			return "", fmt.Errorf("failed to fetch joke")
		}

		// Create a request to pass to the handler
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}

		// Record response
		rec := httptest.NewRecorder()

		// Initialize the handler
		handler := http.HandlerFunc(getRoot)

		// Call the handler
		handler.ServeHTTP(rec, req)

		// Verify status code is 500
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status Internal Server Error; got %v", rec.Code)
		}
	})
}

func TestServerLoad(t *testing.T) {

	const (
		concurrentRequests = 100  // Number of concurrent requests
		totalRequests      = 1000 // Total requests to send
	)

	// WaitGroup to wait for all requests to completed
	var wg sync.WaitGroup

	// Handler for testing
	handler := http.HandlerFunc(getRoot)

	// Channel to collect responses
	responses := make(chan int, totalRequests)

	// Channel to collect any errors
	errors := make(chan error, totalRequests)

	// Semaphore to control the number of concurrent requests
	semaphore := make(chan struct{}, concurrentRequests)

	// make a request to the server for
	for i := 0; i < totalRequests; i++ {
		// Add to the wait group
		wg.Add(1)

		// Acquire a slot in the semaphore
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }() // Release slot in semaphore
			// Create a request
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				errors <- err
				return
			}
			// Record the response
			rec := httptest.NewRecorder()

			// Server the request
			handler.ServeHTTP(rec, req)

			// Send status code to the responses channel
			responses <- rec.Code

			// Check for a 200 OK response
			if rec.Code != http.StatusOK {
				errors <- fmt.Errorf("expected status 200, got %d", rec.Code)
			}
		}()
	}
	// Wait for all requests to complete
	wg.Wait()

	// Close responses and errors channels
	close(responses)
	close(errors)

	// Ensure semaphore is empty
	for i := 0; i < concurrentRequests; i++ {
		semaphore <- struct{}{}
	}
	// Close semaphore channel
	close(semaphore)

	// Check the length of responses to ensure no failures
	if len(responses) < totalRequests {
		t.Errorf("Some requests were unsuccesssful: %d requests made of %v", len(responses), totalRequests)
	}
	// Collect errors
	if len(errors) > 0 {
		t.Errorf("Some requests failed: %d errors", len(errors))
	}
}
