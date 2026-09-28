// Package slack talks to the Slack Web API and OAuth v2.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	apiBase            = "https://slack.com/api"
	oauthAuthorizeBase = "https://slack.com/oauth/v2/authorize"
	defaultBackoff     = time.Second
	maxBackoff         = 30 * time.Second
)

// Channel is a Slack channel.
type Channel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

// Install is the result of a Slack OAuth install.
type Install struct {
	TeamID   string
	TeamName string
	BotToken string
}

// Client posts messages and lists channels for a connected workspace.
type Client interface {
	PostMessage(ctx context.Context, channelID, text string) error
	ListChannels(ctx context.Context) ([]Channel, error)
}

// Installer drives the Slack OAuth install flow.
type Installer interface {
	AuthorizeURL(state string) string
	Exchange(ctx context.Context, code string) (Install, error)
}

// OAuth implements Installer against the Slack API.
type OAuth struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HTTP         *http.Client
}

// NewOAuth builds an OAuth installer.
func NewOAuth(clientID, clientSecret, redirectURL string) *OAuth {
	return &OAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		HTTP:         &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthorizeURL returns the Slack consent URL.
func (o *OAuth) AuthorizeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", o.ClientID)
	q.Set("scope", "chat:write,channels:read,groups:read")
	q.Set("redirect_uri", o.RedirectURL)
	q.Set("state", state)
	return oauthAuthorizeBase + "?" + q.Encode()
}

// Exchange completes the install and returns the bot token and team.
func (o *OAuth) Exchange(ctx context.Context, code string) (Install, error) {
	form := url.Values{}
	form.Set("client_id", o.ClientID)
	form.Set("client_secret", o.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", o.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/oauth.v2.access",
		bytes.NewBufferString(form.Encode()))
	if err != nil {
		return Install{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.HTTP.Do(req)
	if err != nil {
		return Install{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var payload struct {
		OK          bool   `json:"ok"`
		Error       string `json:"error"`
		AccessToken string `json:"access_token"`
		Team        struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Install{}, fmt.Errorf("decode oauth response: %w", err)
	}
	if !payload.OK {
		return Install{}, fmt.Errorf("slack oauth error: %s", payload.Error)
	}
	if payload.AccessToken == "" {
		return Install{}, errors.New("slack oauth response missing access_token")
	}
	return Install{
		TeamID:   payload.Team.ID,
		TeamName: payload.Team.Name,
		BotToken: payload.AccessToken,
	}, nil
}

// APIClient implements Client with a workspace bot token.
type APIClient struct {
	token string
	base  string
	http  *http.Client
}

// NewAPIClient builds a client for a decrypted bot token.
func NewAPIClient(token string) *APIClient {
	return &APIClient{
		token: token,
		base:  apiBase,
		http:  &http.Client{Timeout: 10 * time.Second},
	}
}

// PostMessage posts text to a channel via chat.postMessage.
func (c *APIClient) PostMessage(ctx context.Context, channelID, text string) error {
	payload, _ := json.Marshal(map[string]any{
		"channel": channelID,
		"text":    text,
		"mrkdwn":  true,
	})
	_, err := c.request(ctx, http.MethodPost, c.base+"/chat.postMessage", payload, "application/json")
	return err
}

// ListChannels returns public and private channels the bot can see.
func (c *APIClient) ListChannels(ctx context.Context) ([]Channel, error) {
	var out []Channel
	cursor := ""
	for page := 0; page < 20; page++ {
		q := url.Values{}
		q.Set("types", "public_channel,private_channel")
		q.Set("limit", "200")
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		body, err := c.request(ctx, http.MethodGet, c.base+"/conversations.list?"+q.Encode(), nil, "")
		if err != nil {
			return nil, err
		}
		var payload struct {
			Channels []Channel `json:"channels"`
			Metadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode channels: %w", err)
		}
		out = append(out, payload.Channels...)
		if payload.Metadata.NextCursor == "" {
			break
		}
		cursor = payload.Metadata.NextCursor
	}
	return out, nil
}

// request performs a Slack API call, retrying on rate limits.
func (c *APIClient) request(ctx context.Context, method, endpoint string, body []byte, contentType string) ([]byte, error) {
	backoff := defaultBackoff
	for attempt := 0; attempt < 4; attempt++ {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := retryAfter(resp.Header.Get("Retry-After"), backoff)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			if backoff < maxBackoff {
				backoff *= 2
			}
			continue
		}

		var envelope struct {
			OK    bool   `json:"ok"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			return nil, fmt.Errorf("decode slack response: %w", err)
		}
		if !envelope.OK {
			return nil, fmt.Errorf("slack error: %s", envelope.Error)
		}
		return data, nil
	}
	return nil, errors.New("slack rate limited after retries")
}

func retryAfter(header string, fallback time.Duration) time.Duration {
	if header == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(header)
	if err != nil || seconds < 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
