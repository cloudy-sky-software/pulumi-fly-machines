// Copyright 2022, Cloudy Sky Software LLC.

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudy-sky-software/pulumi-provider-framework/openapi"
	"github.com/cloudy-sky-software/pulumi-provider-framework/state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	testAppName    = "my-app"
	testMachineID  = "3d8d9014b32589"
	testVersion    = "01H3JK5RZXQ8G6YF2N9W4T7B0C"
	testMachineURN = "urn:pulumi:dev::fly-test::fly-machines:apps/v1:Machine::machine"
	testMachineTyp = "fly-machines:apps/v1:Machine"
	testKey1       = "key1"
	testValue1     = "value1"
)

const testCreateJSONPayload = `{
    "appName": "my-app",
    "name": "web",
    "region": "iad",
    "config": {
        "image": "registry-1.docker.io/library/nginx:latest",
        "guest": {
            "cpuKind": "shared",
            "cpus": 1,
            "memoryMb": 256
        }
    }
}
`

const testMachineResponseJSON = `{
    "id": "3d8d9014b32589",
    "name": "web",
    "region": "iad",
    "state": "created",
    "version": "01H3JK5RZXQ8G6YF2N9W4T7B0C",
    "instance_id": "01H3JK5RZXQ8G6YF2N9W4T7B0C",
    "private_ip": "fdaa:0:1:a7b:1::2",
    "config": {
        "image": "registry-1.docker.io/library/nginx:latest",
        "guest": {
            "cpu_kind": "shared",
            "cpus": 1,
            "memory_mb": 256
        }
    }
}
`

func readFileFromProviderResourceDir(t *testing.T, filename string) []byte {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("..", "..", "cmd", "pulumi-resource-fly-machines", filename))
	if err != nil {
		t.Fatalf("Failed reading %s: %v", filename, err)
	}

	return b
}

func makeTestProvider(ctx context.Context, t *testing.T, serverURL string) pulumirpc.ResourceProviderServer {
	t.Helper()

	openapiBytes := readFileFromProviderResourceDir(t, "openapi_generated.yml")
	d := openapi.GetOpenAPISpec(openapiBytes)
	d.Servers[0].URL = serverURL
	openapiBytes, _ = d.MarshalJSON()

	pSchemaBytes := readFileFromProviderResourceDir(t, "schema.json")
	metadataBytes := readFileFromProviderResourceDir(t, "metadata.json")

	p, err := makeProvider(nil, "", "", pSchemaBytes, openapiBytes, metadataBytes)

	if err != nil {
		t.Fatalf("Could not create a provider instance: %v", err)
	}

	_, err = p.Configure(ctx, &pulumirpc.ConfigureRequest{
		Variables:      map[string]string{"fly-machines:config:apiKey": "fakeapikey"},
		SendsOldInputs: true,
	})

	if err != nil {
		t.Fatalf("Error configuring the provider: %v", err)
	}

	return p
}

func testMachineInputs(t *testing.T) map[string]interface{} {
	t.Helper()

	var inputs map[string]interface{}
	if err := json.Unmarshal([]byte(testCreateJSONPayload), &inputs); err != nil {
		t.Fatalf("Failed to unmarshal test payload: %v", err)
	}

	return inputs
}

// testMachineState returns the marshaled state of a Machine
// resource as it would be stored by the engine after create.
func testMachineState(t *testing.T, inputs map[string]interface{}, version string) *structpb.Struct {
	t.Helper()

	outputs := make(map[string]interface{})
	for k, v := range inputs {
		outputs[k] = v
	}
	outputs["id"] = testMachineID
	if version != "" {
		outputs["version"] = version
	}

	s, err := plugin.MarshalProperties(state.GetResourceState(outputs, resource.NewPropertyMapFromMap(inputs)), state.DefaultMarshalOpts)
	require.NoError(t, err)

	return s
}

func marshalInputs(t *testing.T, inputs map[string]interface{}) *structpb.Struct {
	t.Helper()

	s, err := plugin.MarshalProperties(resource.NewPropertyMapFromMap(inputs), state.DefaultMarshalOpts)
	require.NoError(t, err)

	return s
}

func readJSONBody(t *testing.T, body io.Reader) map[string]interface{} {
	t.Helper()

	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))

	return m
}

func TestDiff(t *testing.T) {
	ctx := context.Background()

	p := makeTestProvider(ctx, t, "http://localhost:8080")

	olds := testMachineInputs(t)
	news := testMachineInputs(t)
	news["config"].(map[string]interface{})["image"] = "registry-1.docker.io/library/nginx:1.27"

	resp, err := p.Diff(ctx, &pulumirpc.DiffRequest{
		Id:        testMachineID,
		Urn:       testMachineURN,
		Type:      testMachineTyp,
		Olds:      testMachineState(t, olds, testVersion),
		OldInputs: marshalInputs(t, olds),
		News:      marshalInputs(t, news),
	})
	require.NoError(t, err)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
	assert.Contains(t, resp.Diffs, "config")
	assert.Empty(t, resp.Replaces)
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/apps/"+testAppName+"/machines", r.URL.Path)
		assert.Equal(t, "Bearer fakeapikey", r.Header.Get("Authorization"))

		body := readJSONBody(t, r.Body)
		assert.Equal(t, "web", body["name"])
		assert.Equal(t, "iad", body["region"])
		assert.NotContains(t, body, "app_name", "path params should not be sent in the body")

		config, ok := body["config"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "registry-1.docker.io/library/nginx:latest", config["image"])
		guest, ok := config["guest"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "shared", guest["cpu_kind"])
		assert.EqualValues(t, 256, guest["memory_mb"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testMachineResponseJSON))
	}))
	defer server.Close()

	p := makeTestProvider(ctx, t, server.URL)

	resp, err := p.Create(ctx, &pulumirpc.CreateRequest{
		Urn:        testMachineURN,
		Type:       testMachineTyp,
		Properties: marshalInputs(t, testMachineInputs(t)),
	})
	require.NoError(t, err)
	assert.Equal(t, testMachineID, resp.GetId())

	outputs, err := plugin.UnmarshalProperties(resp.GetProperties(), state.DefaultMarshalOpts)
	require.NoError(t, err)
	assert.Equal(t, testVersion, outputs["version"].StringValue())
	assert.Equal(t, "created", outputs["state"].StringValue())
}

func TestOnPreUpdate(t *testing.T) {
	ctx := context.Background()
	reqURL := "http://localhost:8080/v1/apps/" + testAppName + "/machines/" + testMachineID
	origBody := `{"config":{"image":"registry-1.docker.io/library/nginx:1.27"},"region":"iad"}`

	tests := []struct {
		name            string
		resourceType    string
		version         string
		expectedMethod  string
		expectedVersion string
	}{
		{
			name:            "machine update with version in state",
			resourceType:    testMachineTyp,
			version:         testVersion,
			expectedMethod:  http.MethodPost,
			expectedVersion: testVersion,
		},
		{
			name:           "machine update without version in state",
			resourceType:   testMachineTyp,
			expectedMethod: http.MethodPost,
		},
		{
			name:           "non-machine resource is not modified",
			resourceType:   "fly-machines:apps/v1:Volume",
			version:        testVersion,
			expectedMethod: http.MethodPut,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewBufferString(origBody))
			require.NoError(t, err)

			updateReq := &pulumirpc.UpdateRequest{
				Id:   testMachineID,
				Urn:  testMachineURN,
				Type: tt.resourceType,
				Olds: testMachineState(t, testMachineInputs(t), tt.version),
			}

			p := &flyMachinesProvider{}
			require.NoError(t, p.OnPreUpdate(ctx, updateReq, httpReq))

			assert.Equal(t, tt.expectedMethod, httpReq.Method)
			assert.Equal(t, reqURL, httpReq.URL.String())

			if tt.expectedMethod == http.MethodPut {
				b, err := io.ReadAll(httpReq.Body)
				require.NoError(t, err)
				assert.Equal(t, origBody, string(b))
				return
			}

			getBody, err := httpReq.GetBody()
			require.NoError(t, err)
			bodies := map[string]io.Reader{"Body": httpReq.Body, "GetBody": getBody}

			for name, r := range bodies {
				b, err := io.ReadAll(r)
				require.NoError(t, err, name)
				assert.Equal(t, int64(len(b)), httpReq.ContentLength, name)

				var body map[string]interface{}
				require.NoError(t, json.Unmarshal(b, &body), name)
				assert.Equal(t, "iad", body["region"], name)
				assert.Contains(t, body, "config", name)

				if tt.expectedVersion == "" {
					assert.NotContains(t, body, "current_version", name)
				} else {
					assert.Equal(t, tt.expectedVersion, body["current_version"], name)
				}
			}
		})
	}
}

func TestUpdateMachine(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/apps/"+testAppName+"/machines/"+testMachineID, r.URL.Path)

		body := readJSONBody(t, r.Body)
		assert.Equal(t, testVersion, body["current_version"])
		config, ok := body["config"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "registry-1.docker.io/library/nginx:1.27", config["image"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testMachineResponseJSON))
	}))
	defer server.Close()

	p := makeTestProvider(ctx, t, server.URL)

	olds := testMachineInputs(t)
	news := testMachineInputs(t)
	news["config"].(map[string]interface{})["image"] = "registry-1.docker.io/library/nginx:1.27"

	_, err := p.Update(ctx, &pulumirpc.UpdateRequest{
		Id:        testMachineID,
		Urn:       testMachineURN,
		Type:      testMachineTyp,
		Olds:      testMachineState(t, olds, testVersion),
		OldInputs: marshalInputs(t, olds),
		News:      marshalInputs(t, news),
	})
	require.NoError(t, err)
}

const testMachinesMetadataURN = "urn:pulumi:dev::fly-test::" + machinesMetadataTypeToken + "::metadata"

func testMachinesMetadataInputs(metadata map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"appName":   testAppName,
		"machineId": testMachineID,
		"metadata":  metadata,
	}
}

func TestCreateMachinesMetadata(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/apps/"+testAppName+"/machines/"+testMachineID+"/metadata", r.URL.Path)

		body := readJSONBody(t, r.Body)
		assert.Equal(t, map[string]interface{}{testKey1: testValue1}, body["metadata"])
		assert.NotContains(t, body, "app_name", "path params should not be sent in the body")
		assert.NotContains(t, body, "machine_id", "path params should not be sent in the body")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	p := makeTestProvider(ctx, t, server.URL)

	resp, err := p.Create(ctx, &pulumirpc.CreateRequest{
		Urn:        testMachinesMetadataURN,
		Type:       machinesMetadataTypeToken,
		Properties: marshalInputs(t, testMachinesMetadataInputs(map[string]interface{}{testKey1: testValue1})),
	})
	require.NoError(t, err)
	assert.Equal(t, testMachineID, resp.GetId())
}

func TestUpdateMachinesMetadata(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/apps/"+testAppName+"/machines/"+testMachineID+"/metadata", r.URL.Path)

		body := readJSONBody(t, r.Body)
		assert.Equal(t, map[string]interface{}{testKey1: "value2"}, body["metadata"])
		assert.NotContains(t, body, "app_name", "path params should not be sent in the body")
		assert.NotContains(t, body, "machine_id", "path params should not be sent in the body")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	p := makeTestProvider(ctx, t, server.URL)

	olds := testMachinesMetadataInputs(map[string]interface{}{testKey1: testValue1})
	news := testMachinesMetadataInputs(map[string]interface{}{testKey1: "value2"})

	oldOutputs := make(map[string]interface{})
	for k, v := range olds {
		oldOutputs[k] = v
	}
	oldOutputs["id"] = testMachineID
	oldState, err := plugin.MarshalProperties(state.GetResourceState(oldOutputs, resource.NewPropertyMapFromMap(olds)), state.DefaultMarshalOpts)
	require.NoError(t, err)

	_, err = p.Update(ctx, &pulumirpc.UpdateRequest{
		Id:        testMachineID,
		Urn:       testMachinesMetadataURN,
		Type:      machinesMetadataTypeToken,
		Olds:      oldState,
		OldInputs: marshalInputs(t, olds),
		News:      marshalInputs(t, news),
	})
	require.NoError(t, err)
}

func TestHandleAppIPAssignmentPostCreate(t *testing.T) {
	outputs := handleAppIPAssignmentPostCreate(map[string]interface{}{"ip": nil})
	assert.NotEmpty(t, outputs["id"])
}
