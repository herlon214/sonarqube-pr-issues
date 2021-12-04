package main

import (
	"context"
	"os"
)

func main() {
	// Keys
	apiKey := os.Getenv("SONAR_API_KEY")
	ghToken := os.Getenv("GH_TOKEN")

	ctx := context.Background()

}
