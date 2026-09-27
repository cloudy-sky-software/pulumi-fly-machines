package gen

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

// FixOpenAPIDoc applies patches to the raw OpenAPI spec
// before passing it to pulschema.
func FixOpenAPIDoc(openAPIDoc *openapi3.T) error {
	fixAppEndpoint(openAPIDoc)
	fixUpdateMachineEndpoint(openAPIDoc)
	fixMachineMetadataKeyEndpoint(openAPIDoc)
	fixVolumeEndpoint(openAPIDoc)
	fixPostgresExtensionsEndpoints(openAPIDoc)
	removeDeprecatedVersionProperty(openAPIDoc)

	return nil
}

// fixAppEndpoint renames the app_name path param of the
// /v1/apps/{app_name} endpoint to name so that it matches
// the name property of the App resource. Otherwise the
// provider framework looks for an appName property,
// which the App resource does not have, when it builds
// the read and delete requests.
func fixAppEndpoint(openAPIDoc *openapi3.T) {
	oldPath := "/v1/apps/{app_name}"
	newPath := "/v1/apps/{name}"

	pathItem := openAPIDoc.Paths.Find(oldPath)
	contract.Assertf(pathItem != nil, "Expected to find request path %s", oldPath)

	params := pathItem.Parameters
	for _, op := range pathItem.Operations() {
		params = append(params, op.Parameters...)
	}

	for _, param := range params {
		if param.Value != nil && param.Value.In == "path" && param.Value.Name == "app_name" {
			param.Value.Name = "name"
		}
	}

	openAPIDoc.Paths.Delete(oldPath)
	openAPIDoc.Paths.Set(newPath, pathItem)
}

// fixUpdateMAchineEndpoint adds an update endpoint for Machine resource.
// The API spec uses POST method but we are
// mapping it as a PUT method so that the provider
// framework will accept updates for the resource
// as long as PUT/PATCH is available for a resource.
// In the OnPreUpdate provider callback, we'll modify
// the HTTP request to use the correct method.
func fixUpdateMachineEndpoint(openAPIDoc *openapi3.T) {
	pathItem := openAPIDoc.Paths.Find("/v1/apps/{app_name}/machines/{machine_id}")
	contract.Assertf(pathItem != nil, "Expected to find request path /v1/apps/{app_name}/machines/{machine_id}")

	contract.Assertf(pathItem.Post != nil, "Expected POST operation on /v1/apps/{app_name}/machines/{machine_id}")
	pathItem.Put = pathItem.Post
	pathItem.Put.OperationID = "Update_Machine"
	pathItem.Post = nil
}

func fixMachineMetadataKeyEndpoint(openAPIDoc *openapi3.T) {
	pathItem := openAPIDoc.Paths.Find("/v1/apps/{app_name}/machines/{machine_id}/metadata/{key}")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/machines/{machine_id}/metadata/{key}")

	contract.Assertf(pathItem.Post != nil, "Expected POST operation on /apps/{app_name}/machines/{machine_id}/metadata/{key}")
	pathItem.Post.OperationID = "Machines_create_metadata_key"

	contract.Assertf(pathItem.Delete != nil, "Expected DELETE operation on /apps/{app_name}/machines/{machine_id}/metadata/{key}")
	pathItem.Delete.OperationID = "Machines_delete_metadata_key"
}

func fixVolumeEndpoint(openAPIDoc *openapi3.T) {
	pathItem := openAPIDoc.Paths.Find("/v1/apps/{app_name}/volumes/{volume_id}")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/volumes/{volume_id}")

	contract.Assertf(pathItem.Get != nil, "Expected GET operation on /apps/{app_name}/volumes/{volume_id}")
	pathItem.Get.OperationID = "Volume_get"

	contract.Assertf(pathItem.Put != nil, "Expected PUT operation on /apps/{app_name}/volumes/{volume_id}")
	pathItem.Put.OperationID = "Volume_update"
}

// fixPostgresExtensionsEndpoints updates the operation ID for the
// following endpoints so that they complement each other:
// - /v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions
// - /v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions/{extension_name}
func fixPostgresExtensionsEndpoints(openAPIDoc *openapi3.T) {
	pathItem := openAPIDoc.Paths.Find("/v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions")
	contract.Assertf(pathItem != nil, "Expected to find request path /v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions")

	contract.Assertf(pathItem.Post != nil, "Expected POST operation on /v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions")
	pathItem.Post.OperationID = "Postgres_extensions_create"

	pathItem = openAPIDoc.Paths.Find("/v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions/{extension_name}")
	contract.Assertf(pathItem != nil, "Expected to find request path /v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions/{extension_name}")

	contract.Assertf(pathItem.Delete != nil, "Expected DELETE operation on /v1/postgres/{postgres_cluster_id}/databases/{database_name}/extensions/{extension_name}")
	pathItem.Delete.OperationID = "Postgres_extensions_delete"
}

// removeDeprecatedVersionProperty deletes the deprecated
// Version property from certain schema types that collide
// with the lower-case `version` property. The deprecated
// property causes duplicate class member issues in C#
// classes since class member names use PascalCase per
// convention.
func removeDeprecatedVersionProperty(openAPIDoc *openapi3.T) {
	schemas := []string{"AppSecretsUpdateResp", "DeleteSecretkeyResponse", "DeleteAppSecretResponse", "SetAppSecretResponse", "SetSecretkeyResponse"}
	for _, s := range schemas {
		schema, ok := openAPIDoc.Components.Schemas[s]
		contract.Assertf(ok, "Expected to find schema %s", s)
		delete(schema.Value.Properties, "Version")
	}
}
