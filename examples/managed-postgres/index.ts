import * as pulumi from "@pulumi/pulumi";
import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

import { appName, orgSlug } from "./config.js";
import { db, dbCluster, uuidExtension } from "./db.js";

const app = new fly.apps.v1.App("myapp", {
  name: appName,
  orgSlug,
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
      init: {
        cmd: ["-d", db.name, "-c", "SELECT uuid_generate_v4()"],
      },
      restart: {
        maxRetries: 1,
        policy: "on-failure",
      },
      metadata: {
        fly_platform_version: "v2",
        fly_process_group: "app",
      },
      image: "alpine/psql:latest",
      // Run every 5 minutes.
      schedule: "*/5 * * * *",
    },
    queryParams: {
      delete: {
        // Passes the `force` query param for the delete
        // operation which will force the machine to be
        // deleted even if it is running.
        force: true,
      },
    },
  },
  { dependsOn: [db, uuidExtension] },
);

// Attaching a cluster to an app automatically
// creates a secret called DATABASE_URL.
// The connection string will be for the
// connection pooler (PgBouncer.)
//
// This will trigger a restart of the machines
// in the app.
// https://docs.fly.io/postgres/create-and-connect#attaching-an-app-using-flyctl
//
// Alternatively, if the default secret name
// does not work for you, then create a secret
// manually using the cluster endpoints.
// https://docs.fly.io/postgres/create-and-connect#manually-connect-an-application-using-a-connection-string
//
// Other links:
// https://docs.fly.io/postgres/client-configuration
new fly.postgres.v1.PostgresAttachment(
  "db-app-attachment",
  {
    appName: appName,
    postgresClusterId: dbCluster.id,
  },
  { parent: app, dependsOn: machine },
);

export const appUrl = pulumi.interpolate`https://${appName}.fly.dev`;
export const machineId = machine.id;
