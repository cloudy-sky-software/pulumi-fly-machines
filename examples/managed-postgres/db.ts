import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

import { orgSlug } from "./config.js";

export const dbCluster = new fly.postgres.v1.Postgres("db-cluster", {
  orgSlug,
  plan: "basic",
  region: "ord",
  diskSizeGb: 10,
  pgMajorVersion: "17",
  poolMode: "session",
});

const db = new fly.postgres.v1.PostgresDatabase(
  "db",
  {
    name: "custom-db",
    postgresClusterId: dbCluster.id,
  },
  { parent: dbCluster, dependsOn: dbCluster },
);

// Enable a supported Postgres extension.
// https://docs.fly.io/postgres/extensions#supported-extensions
new fly.postgres.v1.PostgresExtension(
  "pg-cron",
  {
    createSchema: true,
    databaseName: db.name,
    postgresClusterId: dbCluster.id,
    name: "uuid-ossp",
  },
  { parent: db, dependsOn: db },
);
