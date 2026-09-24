package gen

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

// FixOpenAPIDoc applies patches to the raw OpenAPI spec
// before passing it to pulschema.
func FixOpenAPIDoc(openAPIDoc *openapi3.T) error {
	fixMachineMetadataKeyEndpoint(openAPIDoc)
	fixVolumeEndpoint(openAPIDoc)
	fixPostgresExtensionsEndpoints(openAPIDoc)

	return nil
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
