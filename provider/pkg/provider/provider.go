package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	fwCallback "github.com/cloudy-sky-software/pulumi-provider-framework/callback"
	fwRest "github.com/cloudy-sky-software/pulumi-provider-framework/rest"
)

type flyMachinesProvider struct {
	// This embedded struct provides default implementation for callback methods.
	// You are only required to implement the `GetAuthorizationHeader()` method.
	// While not strictly required you might also want to implement `OnConfigure()`.
	fwCallback.UnimplementedProviderCallback

	name    string
	version string

	apiKey string
}

const appIPAssignmentTypeToken = "fly-machines:apps/v1:AppIPAssignment"
const machineTypeToken = "fly-machines:apps/v1:Machine"

var (
	handler  *fwRest.Provider
	callback fwCallback.ProviderCallback
)

func makeProvider(host *provider.HostClient, name, version string, pulumiSchemaBytes, openapiDocBytes, metadataBytes []byte) (pulumirpc.ResourceProviderServer, error) {
	p := &flyMachinesProvider{
		name:    name,
		version: version,
	}

	callback = p
	rp, err := fwRest.MakeProvider(host, name, version, pulumiSchemaBytes, openapiDocBytes, metadataBytes, callback)

	handler = rp.(*fwRest.Provider)

	return rp, err
}

func (p *flyMachinesProvider) GetAuthorizationHeader() string {
	return fmt.Sprintf("%s %s", authSchemePrefix, p.apiKey)
}

// OnConfigure is called by the provider framework when Pulumi calls Configure on
// the resource provider server.
func (p *flyMachinesProvider) OnConfigure(_ context.Context, req *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {
	apiKey, ok := req.GetVariables()["fly-machines:config:apiKey"]
	if !ok {
		// Check if it's set as an env var.
		envVarNames := handler.GetSchemaSpec().Provider.InputProperties["apiKey"].DefaultInfo.Environment
		for _, n := range envVarNames {
			v := os.Getenv(n)
			if v != "" {
				apiKey = v
			}
		}

		// Return an error if the API key is still empty.
		if apiKey == "" {
			return nil, errors.New("api key is required")
		}
	}

	logging.V(3).Info("Configuring FlyMachines API key")
	p.apiKey = apiKey

	return &pulumirpc.ConfigureResponse{
		AcceptSecrets: true,
	}, nil
}

// OnPostCreate is called by the provider framework after the create
// HTTP request succeeds, allowing the outputs to be modified.
func (p *flyMachinesProvider) OnPostCreate(_ context.Context, req *pulumirpc.CreateRequest, outputs interface{}) (map[string]interface{}, error) {
	outputsMap := outputs.(map[string]interface{})

	if fwRest.GetResourceTypeToken(req.GetUrn()) != appIPAssignmentTypeToken {
		return outputsMap, nil
	}

	return handleAppIPAssignmentPostCreate(outputsMap), nil
}

// OnPreUpdate is called by the provider framework before the update
// HTTP request is sent.
func (p *flyMachinesProvider) OnPreUpdate(_ context.Context, updateReq *pulumirpc.UpdateRequest, httpReq *http.Request) error {
	if updateReq.GetType() != machineTypeToken {
		return nil
	}

	return handleUpdateMachineRequest(httpReq, updateReq)
}
