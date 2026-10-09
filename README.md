# Pulumi Native Provider for Fly Machines

[Fly.io](https://fly.io/) runs your apps close to your users on fast-booting VMs. The [Machines API](https://fly.io/docs/machines/) gives you low-level control over Fly Apps, Machines, volumes, IP assignments, secrets and more.

> This provider was generated using [`pulschema`](https://github.com/cloudy-sky-software/pulschema) and [`pulumi-provider-framework`](https://github.com/cloudy-sky-software/pulumi-provider-framework).

## Package SDKs

- Node.js: https://www.npmjs.com/package/@cloudyskysoftware/pulumi-fly-machines
- Python: https://pypi.org/project/pulumi-fly-machines/
- .NET: https://www.nuget.org/packages/Pulumi.FlyMachines
- Go: `import github.com/cloudy-sky-software/pulumi-fly-machines/sdk/go/fly-machines`

## Using The Provider

You'll need a Fly.io API token. Follow Fly's [docs](https://fly.io/docs/machines/api/working-with-machines-api/#api-tokens) for creating one,
for example with `fly tokens create org`.
Then set the token as a secret with `pulumi config set --secret fly-machines:apiKey <token>`.
Alternatively, set the `FLY_MACHINES_APIKEY` environment variable. The provider sends the token as a `Bearer` token.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

const config = new pulumi.Config();
const orgSlug = config.require("orgSlug");

const app = new fly.apps.v1.App("myapp", {
  name: "my-fly-app",
  orgSlug,
});

const ip = new fly.apps.v1.AppIPAssignment("app-ip", {
  appName: app.name,
  type: "shared_v4",
});

const machine = new fly.apps.v1.Machine(
  "myapp-machine",
  {
    appName: app.name,
    config: {
      image: "flyio/hellofly:latest",
      services: [
        {
          ports: [
            { port: 443, handlers: ["tls", "http"] },
            { port: 80, handlers: ["http"] },
          ],
          protocol: "tcp",
          internalPort: 8080,
        },
      ],
    },
  },
  { dependsOn: ip },
);

export const appUrl = pulumi.interpolate`https://${app.name}.fly.dev`;
```

See [`examples/simple`](./examples/simple) for the full example, including how to use `MachinesMetadataKey`
resources to make API-created machines visible to `flyctl`.

### Importing Existing Resources

Import IDs should satisfy all ID segments in the `GET` endpoint for the resource
you are importing. The IDs required in the path should be separated by `/`.
Locate the `GET` endpoint in the [OpenAPI spec](https://github.com/cloudy-sky-software/pulumi-fly-machines/blob/main/provider/cmd/pulumi-gen-fly-machines/openapi.yml).

For example, to read a machine, the path in the OpenAPI spec is: `GET /v1/apps/{app_name}/machines/{machine_id}`.

Thus, the `pulumi import` command to run is:

```bash
# The type fly-machines:apps/v1:Machine can be easily found by using your IDEs
# Go To Definition functionality for the resource and looking at the type
# property defined in the custom resource's class definition.
pulumi import fly-machines:apps/v1:Machine {resourceName} /{app_name}/{machine_id}
```

Alternatively, you can also import using the `import` Pulumi resource option.
Run `pulumi up` to import the resource into your stack's state. Once imported,
you should remove the `import` resource option.

```typescript
const myMachine = new fly.apps.v1.Machine(
  "myMachine",
  { appName: "my-app", config: { image: "flyio/hellofly:latest" } },
  {
    protect: true,
    import: `/my-app/<machine-id>`,
  }
);
```

Refer to the Pulumi [docs](https://www.pulumi.com/docs/iac/adopting-pulumi/import/) for importing a
resource.

## Releasing A New Version

:info: Switch to the `main` branch first and get the latest `git pull origin main && git fetch`. Check what the last release tag was.

1. Regular releases should just increment the patch version unless a minor or a major (breaking changes) version bump is warranted.
1. Update the `CHANGELOG.md` with notes about what will be included in this release.
1. Commit the changelog with `git commit -am "vX.Y.Z"` or something similar and push `git push origin main`.
1. Tag the commit with the release version by running

   ```bash
   git tag vX.Y.Z
   git tag sdk/vX.Y.Z
   ```

1. Push the tags.

   ```bash
   git push --tags
   ```
