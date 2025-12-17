// Package main demonstrates the OAuth2 authentication flow for the freee API.
//
// This example shows how to:
//   - Create an OAuth2 configuration
//   - Generate an authorization URL
//   - Handle the OAuth2 callback
//   - Exchange the authorization code for an access token
//   - Save and load tokens from a file
//
// Usage:
//
//	# Set environment variables
//	export FREEE_CLIENT_ID="your-client-id"
//	export FREEE_CLIENT_SECRET="your-client-secret"
//
//	# Run the example
//	go run main.go
//
// The example will:
//  1. Start a local HTTP server on port 8080
//  2. Print an authorization URL
//  3. Wait for you to visit the URL and authorize the application
//  4. Receive the callback and exchange the code for a token
//  5. Save the token to token.json
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/u-masato/freee-api-go/auth"
)

const (
	// Port for the local callback server.
	callbackPort = "8080"
	// Callback path.
	callbackPath = "/callback"
	// Token file path.
	tokenFile = "token.json"
)

func main() {
	// Get credentials from environment variables
	clientID := os.Getenv("FREEE_CLIENT_ID")
	clientSecret := os.Getenv("FREEE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("FREEE_CLIENT_ID and FREEE_CLIENT_SECRET must be set")
	}

	// Create OAuth2 configuration
	redirectURL := fmt.Sprintf("http://localhost:%s%s", callbackPort, callbackPath)
	config := auth.NewConfig(
		clientID,
		clientSecret,
		redirectURL,
		[]string{"read", "write"},
	)

	// Try to load existing token
	token, err := auth.LoadTokenFromFile(tokenFile)
	if err == nil {
		info := auth.GetTokenInfo(token)
		fmt.Printf("✓ Loaded existing token from %s\n", tokenFile)
		fmt.Printf("  Valid: %v\n", info.Valid)
		fmt.Printf("  Expires in: %v\n", info.ExpiresIn.Round(time.Second))

		if info.Valid {
			fmt.Println("\n✓ Token is still valid. You can use it to make API requests.")
			fmt.Printf("  Access Token: %s...\n", token.AccessToken[:20])
			return
		}

		fmt.Println("\n✗ Token has expired or is invalid.")

		if info.HasRefreshToken {
			fmt.Println("  Attempting to refresh token...")
			ctx := context.Background()
			ts := config.TokenSource(ctx, token)
			newToken, err := ts.Token()
			if err != nil {
				fmt.Printf("  Failed to refresh token: %v\n", err)
				fmt.Println("  Starting new OAuth2 flow...")
			} else {
				fmt.Println("  ✓ Token refreshed successfully")
				if err := auth.SaveTokenToFile(newToken, tokenFile); err != nil {
					log.Printf("Failed to save refreshed token: %v", err)
				} else {
					fmt.Printf("  ✓ Saved refreshed token to %s\n", tokenFile)
				}
				fmt.Printf("  Access Token: %s...\n", newToken.AccessToken[:20])
				return
			}
		}
	} else {
		fmt.Printf("No existing token found. Starting OAuth2 flow...\n\n")
	}

	// Generate a random state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		log.Fatalf("Failed to generate state: %v", err)
	}

	// Create a channel to receive the authorization code
	codeChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Start the callback server
	server := &http.Server{
		Addr:    ":" + callbackPort,
		Handler: callbackHandler(state, codeChan, errorChan),
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Generate and display the authorization URL
	authURL := config.AuthCodeURL(state)
	fmt.Println("Visit this URL to authorize the application:")
	fmt.Printf("\n%s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	// Wait for the callback
	var code string
	select {
	case code = <-codeChan:
		fmt.Println("\n✓ Authorization received")
	case err := <-errorChan:
		log.Fatalf("Authorization failed: %v", err)
	case <-time.After(5 * time.Minute):
		log.Fatal("Authorization timeout (5 minutes)")
	}

	// Shutdown the callback server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Exchange the authorization code for an access token
	fmt.Println("\nExchanging authorization code for access token...")
	token, err = config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Token exchange failed: %v", err)
	}

	fmt.Println("✓ Access token obtained successfully")
	fmt.Printf("  Token Type: %s\n", token.TokenType)
	fmt.Printf("  Access Token: %s...\n", token.AccessToken[:20])
	fmt.Printf("  Expires: %s\n", token.Expiry.Format(time.RFC3339))

	if token.RefreshToken != "" {
		fmt.Printf("  Refresh Token: %s...\n", token.RefreshToken[:20])
	}

	// Save the token to a file
	if err := auth.SaveTokenToFile(token, tokenFile); err != nil {
		log.Printf("Failed to save token: %v", err)
	} else {
		fmt.Printf("\n✓ Token saved to %s\n", tokenFile)
	}

	fmt.Println("\nYou can now use this token to make API requests.")
}

// callbackHandler creates an HTTP handler for the OAuth2 callback.
func callbackHandler(expectedState string, codeChan chan<- string, errorChan chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for error parameter
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			errorChan <- fmt.Errorf("%s: %s", errParam, errDesc)
			http.Error(w, "Authorization failed. You can close this window.", http.StatusBadRequest)
			return
		}

		// Verify state parameter
		state := r.URL.Query().Get("state")
		if state != expectedState {
			errorChan <- fmt.Errorf("state mismatch")
			http.Error(w, "Invalid state parameter. You can close this window.", http.StatusBadRequest)
			return
		}

		// Get the authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			errorChan <- fmt.Errorf("no code in callback")
			http.Error(w, "No authorization code received. You can close this window.", http.StatusBadRequest)
			return
		}

		// Send the code to the main goroutine
		codeChan <- code

		// Display success message
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f5f5f5;
        }
        .container {
            text-align: center;
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 { color: #2ecc71; }
        p { color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1>✓ Authorization Successful</h1>
        <p>You can close this window and return to the terminal.</p>
    </div>
</body>
</html>
`)
	})
}

// generateRandomState generates a random state string for CSRF protection.
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
