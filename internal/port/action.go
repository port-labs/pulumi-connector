package port

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type (
	ActionBody struct {
		Action       string  `json:"action"`
		ResourceType string  `json:"resource_type"`
		Status       string  `json:"status"`
		Trigger      Trigger `json:"trigger"`
		Context      Context `json:"context"`
		Payload      Payload `json:"payload"`
	}

	Trigger struct {
		By     By        `json:"by"`
		At     time.Time `json:"at"`
		Origin string    `json:"origin"`
	}
	Context struct {
		Entity    string `json:"entity,omitempty"`
		Blueprint string `json:"blueprint"`
		RunID     string `json:"runId"`
	}
	Payload struct {
		Entity     Entity         `json:"entity,omitempty"`
		Action     Action         `json:"action,omitempty"`
		Properties map[string]any `json:"properties,omitempty"`
	}

	By struct {
		UserID string `json:"userId"`
		OrgID  string `json:"orgId"`
	}
	Entity struct {
		ID         string         `json:"id,omitempty"`
		Title      string         `json:"title,omitempty"`
		Blueprint  string         `json:"blueprint,omitempty"`
		CreatedAt  time.Time      `json:"createdAt,omitempty"`
		UpdatedAt  time.Time      `json:"updatedAt,omitempty"`
		Properties map[string]any `json:"properties,omitempty"`
		Relations  map[string]any `json:"relations,omitempty"`
		CreatedBy  string         `json:"createdBy,omitempty"`
		UpdatedBy  string         `json:"updatedBy,omitempty"`
	}
	Action struct {
		ID               string           `json:"id"`
		Identifier       string           `json:"identifier"`
		Title            string           `json:"title"`
		UserInputs       UserInputs       `json:"userInputs"`
		InvocationMethod InvocationMethod `json:"invocationMethod"`
		Trigger          string           `json:"trigger"`
		Description      string           `json:"description"`
		Blueprint        string           `json:"blueprint"`
		CreatedAt        string           `json:"createdAt"`
		CreatedBy        string           `json:"createdBy"`
		UpdatedAt        string           `json:"updatedAt"`
		UpdatedBy        string           `json:"updatedBy"`
	}

	InvocationMethod struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}
	UserInputs struct {
		Properties map[string]any `json:"properties"`
		Required   []string       `json:"required"`
	}

	ActionStatus = string
)

const (
	ActionStatusSuccess ActionStatus = "SUCCESS"
	ActionStatusFailure ActionStatus = "FAILURE"
)

// PatchActionRun updates the status of an action run
func (c *Client) PatchActionRun(ctx context.Context, runID string, status ActionStatus) error {
	url := "v1/actions/runs/{run_id}"
	resp, err := c.Client.R().
		SetPathParam("run_id", runID).
		SetContext(ctx).
		SetBody(map[string]string{
			"status": status,
		}).
		Patch(url)
	if err != nil {
		return err
	}
	body := map[string]any{}
	err = json.Unmarshal(resp.Body(), &body)
	if err != nil {
		return err
	}
	if ok, exist := body["ok"]; !exist || !ok.(bool) {
		return errors.New("failed to update action run status")
	}
	return nil
}
