package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	requestTokenURL = "https://api.twitter.com/oauth/request_token"
	callbackURL     = "http://localhost:5173/app/auth/callback/x"
)

type SupportHandler struct{}

func NewSupportHandler() *SupportHandler {
	return &SupportHandler{}
}

func (h *SupportHandler) GetXRequestToken(w http.ResponseWriter, r *http.Request) {
	consumerKey := os.Getenv("X_API_KEY")
	consumerSecret := os.Getenv("X_API_SECRET")

	fmt.Println(consumerKey + ">>>" + consumerSecret)

	// Create OAuth parameters
	params := map[string]string{
		"oauth_callback":         url.QueryEscape(callbackURL),
		"oauth_consumer_key":     consumerKey,
		"oauth_nonce":            uuid.New().String()[:32], // ASCII nonce
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        fmt.Sprintf("%d", time.Now().Unix()),
		"oauth_version":          "1.0",
	}

	// Create signature base string
	baseStr := createSignatureBaseString("POST", requestTokenURL, params)

	// Create signing key
	signingKey := fmt.Sprintf("%s&", url.QueryEscape(consumerSecret))

	// Generate signature
	h1 := hmac.New(sha1.New, []byte(signingKey))
	h1.Write([]byte(baseStr))
	signature := base64.StdEncoding.EncodeToString(h1.Sum(nil))
	params["oauth_signature"] = url.QueryEscape(signature)

	// Create Authorization header
	var headerParts []string
	for key, value := range params {
		headerParts = append(headerParts, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	authHeader := "OAuth " + strings.Join(headerParts, ", ")

	// Create request
	req, err := http.NewRequestWithContext(r.Context(), "POST", requestTokenURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", authHeader)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Parse URL-encoded response
	values, err := url.ParseQuery(string(body))
	if err != nil {
		log.Printf("Error parsing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create JSON response
	jsonResponse := map[string]string{
		"oauth_token":              values.Get("oauth_token"),
		"oauth_token_secret":       values.Get("oauth_token_secret"),
		"oauth_callback_confirmed": values.Get("oauth_callback_confirmed"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func createSignatureBaseString(method, baseURL string, params map[string]string) string {
	var pairs []string
	for key, value := range params {
		if key != "oauth_signature" {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}
	}
	sort.Strings(pairs)
	paramStr := strings.Join(pairs, "&")

	return fmt.Sprintf("%s&%s&%s",
		url.QueryEscape(method),
		url.QueryEscape(baseURL),
		url.QueryEscape(paramStr))
}
