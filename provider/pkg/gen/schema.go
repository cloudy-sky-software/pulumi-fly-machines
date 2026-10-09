// Copyright 2022, Cloudy Sky Software LLC.

package gen

import (
	"bytes"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"

	dotnetgen "github.com/pulumi/pulumi-dotnet/pulumi-language-dotnet/v3/codegen"
	gogen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	nodejsgen "github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	pythongen "github.com/pulumi/pulumi/pkg/v3/codegen/python"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"

	openapigen "github.com/cloudy-sky-software/pulschema/pkg"
	"github.com/cloudy-sky-software/pulschema/pkg/exclusions"

	"github.com/cloudy-sky-software/pulumi-fly-machines/provider/pkg/gen/examples"
)

const packageName = "fly-machines"

// PulumiSchema will generate a Pulumi schema for the given k8s schema.
func PulumiSchema(openapiDoc openapi3.T) (pschema.PackageSpec, openapigen.ProviderMetadata, openapi3.T) {
	pkg := pschema.PackageSpec{
		Name:        packageName,
		Description: "A Pulumi package for creating and managing FlyMachines resources.",
		DisplayName: "FlyMachines",
		License:     "Apache-2.0",
		Keywords: []string{
			"pulumi",
			packageName,
			"category/cloud",
			"kind/native",
		},
		Homepage:   "https://cloudysky.software",
		Publisher:  "Cloudy Sky Software",
		Repository: "https://github.com/cloudy-sky-software/pulumi-fly-machines",

		Config: pschema.ConfigSpec{
			Variables: map[string]pschema.PropertySpec{
				"apiKey": {
					Description: "The API key",
					TypeSpec:    pschema.TypeSpec{Type: "string"},
					Language: map[string]pschema.RawMessage{
						"csharp": rawMessage(dotnetgen.CSharpPropertyInfo{
							Name: "ApiKey",
						}),
					},
					Secret: true,
				},
			},
		},

		Provider: &pschema.ResourceSpec{
			ObjectTypeSpec: pschema.ObjectTypeSpec{
				Description: "The provider type for the FlyMachines package.",
				Type:        "object",
			},
			InputProperties: map[string]pschema.PropertySpec{
				"apiKey": {
					DefaultInfo: &pschema.DefaultSpec{
						Environment: []string{
							"FLY_MACHINES_APIKEY",
						},
					},
					Description: "The FlyMachines API key.",
					TypeSpec:    pschema.TypeSpec{Type: "string"},
					Language: map[string]pschema.RawMessage{
						"csharp": rawMessage(dotnetgen.CSharpPropertyInfo{
							Name: "ApiKey",
						}),
					},
					Secret: true,
				},
			},
		},

		PluginDownloadURL: "github://api.github.com/cloudy-sky-software/pulumi-fly-machines",
		Types:             map[string]pschema.ComplexTypeSpec{},
		Resources:         map[string]pschema.ResourceSpec{},
		Functions:         map[string]pschema.FunctionSpec{},
		Language:          map[string]pschema.RawMessage{},
	}

	csharpNamespaces := map[string]string{
		"fly-machines": "FlyMachines",
		// TODO: Is this needed?
		"": "Provider",
	}

	openAPICtx := &openapigen.OpenAPIContext{
		Doc: openapiDoc,
		Pkg: &pkg,
		Exclusions: []exclusions.Exclusion{
			// DELETE /v1/apps/{app_name}/certificates/{hostname} simply
			// allows deletion of any type of certificate -- Acme or custom.
			// Those resource types have their own DELETE endpoints already.
			{
				Method:      "DELETE",
				PathPattern: "/v1/apps/{app_name}/certificates/{hostname}",
				PatternType: exclusions.PatternTypeExact,
			},
		},
	}

	providerMetadata, updatedOpenAPIDoc, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	if err != nil {
		contract.Failf("generating resources from OpenAPI spec: %v", err)
	}

	// Add examples to resources
	for k, v := range examples.ResourceExample {
		if r, ok := pkg.Resources[k]; ok {
			r.Description += "\n\n" + v
			pkg.Resources[k] = r
		}
	}

	pkg.Language["csharp"] = rawMessage(dotnetgen.CSharpPackageInfo{
		RootNamespace: "Pulumi",
		PackageReferences: map[string]string{
			"Pulumi": "3.*",
		},
		Namespaces: csharpNamespaces,
	})

	pkg.Language["go"] = rawMessage(gogen.GoPackageInfo{
		ImportBasePath: "github.com/cloudy-sky-software/pulumi-fly-machines/sdk/go/fly-machines",
	})
	pkg.Language["nodejs"] = rawMessage(nodejsgen.NodePackageInfo{
		PackageName: "@cloudyskysoftware/pulumi-fly-machines",
	})
	pkg.Language["python"] = rawMessage(pythongen.PackageInfo{
		PackageName: "pulumi_fly_machines",
		Requires: map[string]string{
			"pulumi": ">=3.0.0,<4.0.0",
		},
		PyProject: struct {
			Enabled bool `json:"enabled,omitempty"`
		}{Enabled: true},
	})

	// Add the "Get certificate details" endpoint to the
	// /custom and /acme certificate endpoint resources.
	providerMetadata.ResourceCRUDMap["fly-machines:apps/v1:AppCertificatesCustom"].R = new(`/v1/apps/{app_name}/certificates/{hostname}`)
	providerMetadata.ResourceCRUDMap["fly-machines:apps/v1:AppCertificatesAcme"].R = new(`/v1/apps/{app_name}/certificates/{hostname}`)

	metadata := openapigen.ProviderMetadata{
		ResourceCRUDMap:  providerMetadata.ResourceCRUDMap,
		AutoNameMap:      providerMetadata.AutoNameMap,
		SDKToAPINameMap:  providerMetadata.SDKToAPINameMap,
		APIToSDKNameMap:  providerMetadata.APIToSDKNameMap,
		PathParamNameMap: providerMetadata.PathParamNameMap,
	}
	return pkg, metadata, updatedOpenAPIDoc
}

func rawMessage(v interface{}) pschema.RawMessage {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	contract.Assert(err == nil)
	return out.Bytes()
}
