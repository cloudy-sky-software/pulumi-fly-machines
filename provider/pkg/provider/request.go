// Copyright 2022, Cloudy Sky Software LLC.

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudy-sky-software/pulumi-provider-framework/state"
	"github.com/google/uuid"
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

// handleCreateMachinesMetadataRequest handles create
// requests for machine metadata.
//
// The original API spec does not have a POST endpoint
// for machine metadata. A fake one, matching the PATCH
// endpoint, was added to the spec so that the CRUD
// metadata would have a create endpoint for the
// MachinesMetadata resource.
//
// The request is modified in-place to use the PATCH method.
func handleCreateMachinesMetadataRequest(httpReq *http.Request) {
	httpReq.Method = http.MethodPatch
}

// handleAppIPAssignmentPostCreate sets a pseudo id
// for the AppIPAssignment resource since the API
// response does not include an id. The ip property
// can't be used because it is null for egress-pair
// assignments, which set ip_pair instead.
func handleAppIPAssignmentPostCreate(outputs map[string]interface{}) map[string]interface{} {
	outputs["id"] = uuid.NewString()
	return outputs
}

// handleMachinesMetadataPostCreate sets the id
// for the MachinesMetadata resource to the id
// of the machine that the metadata belongs to.
// The create request (a PATCH) returns 204 No Content,
// so there is no response body to get the id from.
func handleMachinesMetadataPostCreate(req *pulumirpc.CreateRequest, outputs map[string]interface{}) (map[string]interface{}, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), state.DefaultUnmarshalOpts)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal inputs")
	}

	// fly.io uses snake_case in its responses but we are
	// looking for machineId from the inputs which uses
	// camelCase.
	machineID, ok := inputs["machineId"]
	if !ok || !machineID.IsString() || machineID.StringValue() == "" {
		return nil, errors.New("machineId input is required to set the id for the machine metadata resource")
	}

	outputs["id"] = machineID.StringValue()
	return outputs, nil
}
