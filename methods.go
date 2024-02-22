package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Function to get coordinates from address using the Google Maps Geocoding API
func getCoordinatesFromAddress(address string) (googleMapsCoordinates, error) {
	apiKey := "AIzaSyBznSC8S1mPU-GPjsxuagQqnNK3a8xVOl4"
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?key=%s&place_id=%s", apiKey, address)
	// Make HTTP request
	response, err := http.Get(url)
	if err != nil {
		return googleMapsCoordinates{}, err
	}
	defer response.Body.Close()

	// Read and parse the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return googleMapsCoordinates{}, err
	}

	// Parse the JSON response
	var result googleMapsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return googleMapsCoordinates{}, err
	}

	// Check if the response status is OK
	if result.Status != "OK" || len(result.Results) == 0 {
		return googleMapsCoordinates{}, fmt.Errorf("Geocoding API request failed")
	}

	// Extract coordinates from the first result
	coordinates := result.Results[0].Geometry.Location
	return coordinates, nil
}
