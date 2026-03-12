package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zCyberSecurity/zapi/internal/model"
	"gorm.io/gorm"
)

type Proxy struct {
	db     *gorm.DB
	client *http.Client
}

func New(db *gorm.DB) *Proxy {
	return &Proxy{
		db:     db,
		client: &http.Client{Timeout: 180 * time.Second},
	}
}

// FindProvider locates the enabled provider that serves modelID.
func (p *Proxy) FindProvider(modelID string) (*model.Provider, *model.ProviderModel, error) {
	var pm model.ProviderModel
	err := p.db.Preload("Provider").
		Joins("JOIN providers ON providers.id = provider_models.provider_id AND providers.enabled = true AND providers.deleted_at IS NULL").
		Where("provider_models.model_id = ? AND provider_models.enabled = ? AND provider_models.deleted_at IS NULL", modelID, true).
		First(&pm).Error
	if err != nil {
		return nil, nil, fmt.Errorf("model %q not found or not enabled", modelID)
	}
	return &pm.Provider, &pm, nil
}

// ForwardStream sends body to the upstream, streams the response into w,
// and returns extracted token usage (prompt, completion, total).
func (p *Proxy) ForwardStream(w http.ResponseWriter, origReq *http.Request, provider *model.Provider, path string, body []byte) (int, int, int, error) {
	url := strings.TrimRight(provider.BaseURL, "/") + path
	log.Printf("[DEBUG] ForwardStream upstream=%s provider=%s model=%s", url, provider.Name, ParseOpenAIRequest(body).Model)

	req, err := http.NewRequestWithContext(origReq.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	flusher, canFlush := w.(http.Flusher)
	var prompt, completion, total int

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		if canFlush {
			flusher.Flush()
		}
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}
			p, c, t := ExtractOpenAIUsage([]byte(data))
			if t > 0 {
				prompt, completion, total = p, c, t
			}
		}
	}

	return prompt, completion, total, scanner.Err()
}

// ForwardBuffered sends body to the upstream and returns the full response body.
func (p *Proxy) ForwardBuffered(origReq *http.Request, provider *model.Provider, path string, body []byte) (int, []byte, error) {
	url := strings.TrimRight(provider.BaseURL, "/") + path
	log.Printf("[DEBUG] ForwardBuffered upstream=%s provider=%s model=%s", url, provider.Name, ParseOpenAIRequest(body).Model)

	req, err := http.NewRequestWithContext(origReq.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, err
}

func copyHeaders(dst, src http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

// --- helpers used by handlers ---

type openAIRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

func ParseOpenAIRequest(body []byte) openAIRequest {
	var r openAIRequest
	json.Unmarshal(body, &r)
	return r
}

type openAIUsage struct {
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func ExtractOpenAIUsage(body []byte) (prompt, completion, total int) {
	var r openAIUsage
	json.Unmarshal(body, &r)
	return r.Usage.PromptTokens, r.Usage.CompletionTokens, r.Usage.TotalTokens
}

// ReplaceModel rewrites the "model" field in a JSON body.
func ReplaceModel(body []byte, newModel string) []byte {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return body
	}
	b, _ := json.Marshal(newModel)
	m["model"] = b
	out, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return out
}

// InjectStreamOptions adds stream_options.include_usage=true so providers return usage in the last SSE chunk.
func InjectStreamOptions(body []byte) []byte {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return body
	}
	m["stream_options"] = json.RawMessage(`{"include_usage":true}`)
	out, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return out
}
