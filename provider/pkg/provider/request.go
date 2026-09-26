// Copyright 2022, Cloudy Sky Software LLC.

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudy-sky-software/pulumi-provider-framework/state"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
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
// The request is modified in-place to use the POST
// method. If currentVersion is not empty, the request
// body will also carry it as the current_version property.
func handleUpdateMachineRequest(httpReq *http.Request, updateReq *pulumirpc.UpdateRequest) error {
	body, err := httpReq.GetBody()
	if err != nil {
		return fmt.Errorf("get request body: %w", err)
	}

	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("reading request body: %w", err)
	}

	var reqBody map[string]any
	if err := json.Unmarshal(b, &reqBody); err != nil {
		return fmt.Errorf("unmarshal request body: %w", err)
	}

	olds, err := plugin.UnmarshalProperties(updateReq.GetOlds(), state.DefaultMarshalOpts)
	if err != nil {
		return errors.Wrap(err, "unmarshal olds")
	}

	var currentVersion string
	if v, ok := olds["version"]; ok && v.IsString() {
		currentVersion = v.StringValue()
	}

	if currentVersion != "" {
		reqBody["current_version"] = currentVersion
	}

	updatedBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling modified request body: %w", err)
	}

	httpReq.Method = http.MethodPost
	httpReq.ContentLength = int64(len(updatedBody))
	httpReq.Body = io.NopCloser(bytes.NewReader(updatedBody))
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(updatedBody)), nil
	}

	return nil
}
