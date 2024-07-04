package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"github.com/wlcmtunknwndth/FCSxVK/internal/proxy"
	"google.golang.org/api/option"
	"net/http"
	"strings"
)

const scope = "internal.AI.gemini."

type Gemini struct {
	model  *genai.GenerativeModel
	client *genai.Client
}

func New(ctx context.Context, apiKey, proxyUrl, username, password string) (*Gemini, error) {
	const op = scope + "New"

	//client, err := genai.NewClient(ctx)
	c := &http.Client{
		Transport: &proxy.APIKeyProxyTransport{
			APIKey:    apiKey,
			Username:  username,
			Password:  password,
			Transport: nil,
			ProxyURL:  proxyUrl,
		},
	}
	client, err := genai.NewClient(ctx, option.WithHTTPClient(c), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")

	return &Gemini{
		model:  model,
		client: client,
	}, nil
}

func (g *Gemini) HandlePrompt(ctx context.Context, msg string) (string, error) {
	const op = scope + "HandlePrompt"

	resp, err := g.model.GenerateContent(ctx, genai.Text(msg))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	txt, err := retrieveText(resp)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return txt, nil
}

type Content struct {
	Parts []string `json:"Parts"`
	Role  string   `json:"Role"`
}
type Candidates struct {
	Content *Content `json:"Content"`
}
type ContentResponse struct {
	Candidates *[]Candidates `json:"Candidates"`
}

func retrieveText(resp *genai.GenerateContentResponse) (string, error) {
	const op = scope + "retrieveText"

	marshalResponse, err := json.MarshalIndent(resp, "", " ")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var generateResponse ContentResponse
	if err = json.Unmarshal(marshalResponse, &generateResponse); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	var builder strings.Builder
	err = nil
	for _, cad := range *generateResponse.Candidates {
		if cad.Content != nil {
			for _, part := range cad.Content.Parts {
				if _, err = builder.WriteString(part); err != nil {
					err = errors.Join(err)
					continue
				}
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return builder.String(), nil
}

func (g *Gemini) Close() error {
	const op = scope + "Scope"
	if err := g.client.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}