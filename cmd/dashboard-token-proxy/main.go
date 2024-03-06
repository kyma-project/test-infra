package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type GithubResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type DashboardResponse struct {
	AccessToken string `json:"access_token"`
	Success     bool   `json:"success"`
}

func main() {
	clientID, clientSecret, ghURL, err := getEnvSecrets()
	if err != nil {
		log.Fatalf("Failed to get secrets value: %s", err)
	}

	authorizationURL := fmt.Sprintf("https://%s/login/oauth/access_token", ghURL)

	http.HandleFunc("/", HandleTokenRequest(clientID, clientSecret, ghURL, authorizationURL))
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func HandleTokenRequest(clientID, clientSecret, ghURL, authorizationURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if clientID == "" || clientSecret == "" {
			log.Printf("Client data is missing")
			http.Error(w, "Client data is missing", http.StatusBadRequest)
			return
		}

		client := &http.Client{}

		err := r.ParseForm()
		if err != nil {
			log.Printf("failed to parse form: %v", err)
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		code := r.Form["oauthz_code"][0]

		data := url.Values{}
		data.Add("code", code)
		data.Add("client_id", clientID)
		data.Add("client_secret", clientSecret)

		req, err := http.NewRequest("POST", authorizationURL, strings.NewReader(data.Encode()))
		if err != nil {
			log.Printf("could not create Github request: %v", err)
			http.Error(w, "Could not create login request", http.StatusBadRequest)
			return
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("failed to login to Github: %v", err)
			http.Error(w, "Could not login to Github", http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed to read response body: %v", err)
			http.Error(w, "failed to read response body", http.StatusBadRequest)
			return
		}
		var responseData GithubResponse
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			log.Printf("failed to unmarshal response body: %v", err)
			http.Error(w, "failed to unmarshal response body", http.StatusBadRequest)
			return
		}

		dashboardResponse := DashboardResponse{
			AccessToken: responseData.AccessToken,
			Success:     responseData.AccessToken != "",
		}

		b, err := json.Marshal(dashboardResponse)
		if err != nil {
			log.Printf("failed to marshal response body: %v", err)
			http.Error(w, "failed to marshal response body", http.StatusBadRequest)
			return
		}

		w.Header().Add("content-type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", fmt.Sprintf("https://pages.%s", ghURL))
		// TODO: The web-application does not define an HSTS header, leaving it vulnerable to attack.
		w.Write(b)
	}
}

func getEnvSecrets() (string, string, string, error) {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	ghURL := os.Getenv("GH_BASE_URL")

	if clientID == "" {
		return "", "", "", fmt.Errorf("client id is required")
	}

	if clientSecret == "" {
		return "", "", "", fmt.Errorf("client secret is required")
	}

	if ghURL == "" {
		return "", "", "", fmt.Errorf("github url is required")
	}

	return clientID, clientSecret, ghURL, nil
}
