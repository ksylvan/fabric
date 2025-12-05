package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	debuglog "github.com/danielmiessler/fabric/internal/log"
	"github.com/danielmiessler/fabric/internal/util"
)

const (
	tokenEndpoint = "https://login.microsoftonline.com"
	defaultScope  = "https://graph.microsoft.com/.default"
)

// TokenResponse represents Microsoft Graph token response
type TokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

// authenticateWithClientSecret performs app-only authentication using client credentials flow
func (c *Client) authenticateWithClientSecret() error {
	if c.ClientID.Value == "" || c.ClientSecret.Value == "" {
		return fmt.Errorf("client ID and client secret are required for app-only authentication")
	}

	// Determine tenant ID
	tenantID := c.TenantID.Value
	if tenantID == "" {
		tenantID = "common" // Default to common endpoint
	}

	// Build token request URL
	tokenURL := fmt.Sprintf("%s/%s/oauth2/v2.0/token", tokenEndpoint, tenantID)

	// Build request payload
	payload := url.Values{}
	payload.Set("client_id", c.ClientID.Value)
	payload.Set("client_secret", c.ClientSecret.Value)
	payload.Set("scope", defaultScope)
	payload.Set("grant_type", "client_credentials")

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", tokenURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("no access token received")
	}

	c.accessToken = tokenResp.AccessToken

	// Initialize conversation manager after successful authentication
	// This will be called from the main client
	return nil
}

// authenticateWithOAuth performs OAuth authentication flow
func (c *Client) authenticateWithOAuth() error {
	storage, err := util.NewOAuthStorage()
	if err != nil {
		return fmt.Errorf("failed to create OAuth storage: %w", err)
	}

	// Check if we have a valid token
	if storage.HasValidToken("copilot", 5) {
		token, err := storage.LoadToken("copilot")
		if err == nil && token != nil && token.AccessToken != "" {
			c.accessToken = token.AccessToken
			return nil
		}
	}

	// Start OAuth flow
	debuglog.Debug(debuglog.Basic, "Starting Microsoft Graph OAuth authentication...")
	token, err := c.runOAuthFlow()
	if err != nil {
		return fmt.Errorf("OAuth flow failed: %w", err)
	}

	// Store token
	oauthToken := &util.OAuthToken{
		AccessToken: token,
		ExpiresAt:   time.Now().Add(60 * time.Minute).Unix(), // 1 hour expiry
		TokenType:   "Bearer",
	}
	if err := storage.SaveToken("copilot", oauthToken); err != nil {
		return fmt.Errorf("failed to store OAuth token: %w", err)
	}

	c.accessToken = token
	return nil
}

// runOAuthFlow initiates OAuth authentication flow
func (c *Client) runOAuthFlow() (string, error) {
	// TODO: Implement OAuth flow
	// This would involve:
	// 1. Redirecting user to Microsoft login
	// 2. Handling callback with authorization code
	// 3. Exchanging code for access token
	// 4. Returning the access token

	// For now, return error as OAuth is not fully implemented
	return "", fmt.Errorf("OAuth flow not yet implemented - please use client secret authentication")
}
