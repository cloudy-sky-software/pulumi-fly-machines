import * as pulumi from "@pulumi/pulumi";
import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

const app = new fly.apps.v1.App("myapp", {
  name: "pulumi-fly-machines-test-app",
  orgSlug: "ENTER_ORG_SLUG_HERE",
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
      // NOTE: Metadata can be set using the metadata
      // property or as individual `MachinesMetadataKey`
      // resources (see below.)
      //
      // If you use both approaches, be sure to
      // add `metadata` property to `ignoreChanges`,
      // so that Pulumi doesn't show a drift in the
      // machine's config due to new metadata keys
      // being added separately. You don't need to
      // do this, of course, if you only use one of
      // the methods to manage a machine's metadata.
      //
      // metadata: {
      //   "key1": "value1",
      // },
      image: "flyio/hellofly:latest",
      services: [{}],
    },
  },
  {
    dependsOn: ip,
    // See above note about the metadata property.
    // ignoreChanges: ["config.metadata"]
  },
);

// Make the machine visible to flyctl.
// If you don't want this machine to be
// visible to flyctl then remove these.
//
// https://docs.fly.io/machines/guides-examples/managing-machines-with-the-api#make-api-created-machines-visible-to-flyctl
new fly.apps.v1.MachinesMetadataKey("machine-metadata-platform-version", {
  appName: app.name,
  key: "fly_platform_version",
  machineId: machine.id,
  value: "v2",
});

new fly.apps.v1.MachinesMetadataKey("machine-metadata-process-group", {
  appName: app.name,
  key: "fly_process_group",
  machineId: machine.id,
  value: "app",
});

export const appName = app.name;
export const appUrl = pulumi.interpolate`https://${appName}.fly.dev`;
export const machineId = machine.id;
