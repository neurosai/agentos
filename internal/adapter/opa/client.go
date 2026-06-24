package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/neurosai/agentos/internal/domain/policy"
)

// Client evaluates Rego policies via OPA REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type evalResponse struct {
	Result json.RawMessage `json:"result"`
}

// Evaluate runs OPA data API for the given package and maps a Decision.
func (c *Client) Evaluate(ctx context.Context, pkg string, input policy.EvaluationInput) (policy.Decision, error) {
	path := packageToPath(pkg)
	body, err := json.Marshal(map[string]any{"input": inputToMap(input)})
	if err != nil {
		return policy.Decision{}, err
	}
	allow, err := c.queryBool(ctx, path+"/allow", body)
	if err != nil {
		return policy.Decision{}, err
	}
	requireApproval, _ := c.queryBool(ctx, path+"/require_approval", body)
	denyReason, _ := c.queryString(ctx, path+"/deny_reason", body)

	dec := policy.Decision{EvaluatedAt: time.Now().UTC()}
	switch {
	case denyReason != "":
		dec.Effect = policy.EffectDeny
		dec.DenyReason = denyReason
	case requireApproval:
		dec.Effect = policy.EffectRequireApproval
	case allow:
		dec.Effect = policy.EffectAllow
	default:
		dec.Effect = policy.EffectDeny
		dec.DenyReason = "policy denied"
	}
	return dec, nil
}

func (c *Client) queryBool(ctx context.Context, path string, body []byte) (bool, error) {
	raw, err := c.post(ctx, path, body)
	if err != nil {
		return false, err
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err != nil {
		return false, nil
	}
	return v, nil
}

func (c *Client) queryString(ctx context.Context, path string, body []byte) (string, error) {
	raw, err := c.post(ctx, path, body)
	if err != nil {
		return "", err
	}
	var v string
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", nil
	}
	return v, nil
}

func (c *Client) post(ctx context.Context, path string, body []byte) (json.RawMessage, error) {
	url := c.baseURL + "/v1/data/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("opa %s: %s", url, string(data))
	}
	var out evalResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out.Result, nil
}

func packageToPath(pkg string) string {
	pkg = strings.TrimPrefix(pkg, "agentos.")
	if pkg == "" {
		return "agentos"
	}
	return "agentos/" + strings.ReplaceAll(pkg, ".", "/")
}

func inputToMap(in policy.EvaluationInput) map[string]any {
	m := map[string]any{
		"requestId": in.RequestID,
		"action":    in.Action,
		"subject": map[string]any{
			"type":      in.Subject.Type,
			"id":        in.Subject.ID,
			"roles":     in.Subject.Roles,
			"groups":    in.Subject.Groups,
			"group_ids": in.Subject.Groups,
		},
		"actingAgent": map[string]any{
			"id":     in.ActingAgent.ID,
			"labels": in.ActingAgent.Labels,
		},
		"resource": map[string]any{
			"type":      in.Resource.Type,
			"id":        in.Resource.ID,
			"tenantId":  in.Resource.TenantID,
			"risk":      in.Resource.Risk,
			"namespace": in.Resource.Namespace,
		},
		"context": map[string]any{
			"taskId":         in.Context.TaskID,
			"classification": in.Context.Classification,
			"workspace":      in.Context.Workspace,
			"sourceTrust":    in.Context.SourceTrust,
			"environment":    in.Context.Environment,
		},
	}
	if in.Record.Classification != "" || in.Record.Namespace != "" {
		m["record"] = map[string]any{
			"classification": in.Record.Classification,
			"namespace":      in.Record.Namespace,
		}
	}
	return m
}

// PackageForAction maps action prefix to OPA package name.
func PackageForAction(action string) string {
	switch action {
	case "invoke":
		return "agentos.tools"
	case "read", "write", "delete", "search":
		return "agentos.memory"
	default:
		if strings.HasPrefix(action, "task.") {
			return "agentos.tasks"
		}
		return "agentos.tasks"
	}
}
