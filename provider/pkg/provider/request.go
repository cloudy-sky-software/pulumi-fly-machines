// Copyright 2022, Cloudy Sky Software LLC.

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	authSchemePrefix = "Bearer"
)

// handleUpdateMachineRequest handles update
// requests for machines.
//
// The original endpoint uses a POST method.
// It was overridden in the spec as a PUT
// endpoint so that the CRUD metadata would
// have an update endpoint for the Machine resource.
//
// The request body must also carry a current_version
// property.
func handleUpdateMachineRequest(ctx context.Context, httpReq *http.Request) error {
	body, err := httpReq.GetBody()
	if err != nil {
		return fmt.Errorf("get request body: %w", err)
	}

	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("reading request body: %w", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(b, reqBody); err != nil {
		return fmt.Errorf("unmarshal request body: %w", err)
	}

	if version, ok := reqBody["version"]; ok {
		reqBody["current_version"] = version
	}

	updatedBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling modified request body: %w", err)
	}

	buf := bytes.NewBuffer(updatedBody)

	// Create a new request.
	httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, httpReq.URL.String(), buf)
	if err != nil {
		return fmt.Errorf("initializing modified request: %w", err)
	}

	return nil
}
