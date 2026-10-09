import * as pulumi from "@pulumi/pulumi";
import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

const config = new pulumi.Config();
const orgSlug = config.require("orgSlug");

const app = new fly.apps.v1.App("myapp", {
  name: "pulumi-fly-machines-test-app",
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
      guest: {
        cpuKind: "shared",
        cpus: 1,
        memoryMb: 256,
      },
      init: {},
      restart: {
        maxRetries: 10,
        policy: "on-failure",
      },
      // NOTE: Metadata can be set using the metadata
      // property or as individual `MachinesMetadataKey`
      // resources or a `MachinesMetadata` resource
      // (see below.)
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
      services: [
        {
          ports: [
            { port: 443, handlers: ["tls", "http"] },
            { port: 80, handlers: ["http"] },
          ],
          protocol: "tcp",
          internalPort: 8080,
          autostop: "suspend",
          autostart: true,
          minMachinesRunning: 1,
        },
      ],
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
new fly.apps.v1.MachinesMetadata("machine-metadata", {
  appName: app.name,
  machineId: machine.id,
  metadata: {
    fly_platform_version: "v2",
    fly_process_group: "app",
    testkey: "value",
  },
});

export const appName = app.name;
export const appUrl = pulumi.interpolate`https://${appName}.fly.dev`;
export const machineId = machine.id;
