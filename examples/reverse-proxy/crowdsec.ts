import { readFileSync } from "fs";
import path from "path";
import { cwd } from "process";

import * as fly from "@cloudyskysoftware/pulumi-fly-machines";
import * as random from "@pulumi/random";

import { crowdSecEnrollKey, flyOrgSlug, region } from "./config.js";

import { FlyApp } from "./flyApp.js";

// 8080 is the default port for the image `crowdsecurity/crowdsec`.
const crowdSecPort = 8080;
const crowdSecAppName = "crowdsec";

const crowdSecApp = new FlyApp(crowdSecAppName, {
  ipAssignments: ["private_v6"],
  orgSlug: flyOrgSlug,
});

const crowdSecDataFork1 = new fly.apps.v1.Volume(
  "crowdsec-volume-fork-1",
  {
    appName: crowdSecAppName,
    autoBackupEnabled: true,
    encrypted: true,
    name: "data",
    region,
  },
  { protect: true },
);

export const openRestyApiKey = new random.RandomBytes("openresty-api-key", {
  length: 16,
  keepers: {
    version: "1",
  },
}).hex;

const secrets = new fly.apps.v1.Secrets("crowdsec-secrets", {
  appName: crowdSecAppName,
  values: {
    BOUNCER_KEY_openresty: openRestyApiKey,
    ENROLL_KEY: crowdSecEnrollKey,
  },
});

const crowdSecAcquisitionsConfig = Buffer.from(
  readFileSync(
    path.join(
      cwd().endsWith("infra") ? "." : path.join("src", "infra"),
      "configs",
      "crowdsec",
      "acquis.yaml",
    ),
    {
      encoding: "utf-8",
    },
  ),
  "utf-8",
).toString("base64");

const crowdSecConfigOverride = Buffer.from(
  readFileSync(
    path.join(
      cwd().endsWith("infra") ? "." : path.join("src", "infra"),
      "configs",
      "crowdsec",
      "config.yaml",
    ),
    {
      encoding: "utf-8",
    },
  ),
  "utf-8",
).toString("base64");

export const crowdSecMachine = new fly.apps.v1.Machine(
  "crowdsec-machine",
  {
    appName: crowdSecAppName,
    region: crowdSecDataFork1.region,
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
      metadata: {
        fly_platform_version: "v2",
        fly_process_group: "crowdsec",
      },
      cacheDisk: {
        path: "/etc/crowdsec",
        sizeMb: 250,
      },
      services: [
        {
          ports: [{ port: 80, handlers: ["http"] }],
          protocol: "tcp",
          internalPort: crowdSecPort,
          autostop: "off",
          autostart: true,
          minMachinesRunning: 1,
        },
      ],
      containers: [
        {
          name: "crowdsec-engine",
          image: "crowdsecurity/crowdsec:v1.8.2",
          env: {
            // https://app.crowdsec.net/hub/author/crowdsecurity/collections/nginx
            COLLECTIONS: "crowdsecurity/nginx",
            CONFIG_FILE: "/crowdsec-config/config.yaml",
            ENROLL_TAGS: "deskypus",
            TZ: "America/Los_Angeles",
            USE_WAL: "true",
          },
          healthchecks: [
            {
              exec: {
                command: ["cscli", "lapi", "status"],
              },
              interval: 30,
              timeout: 5,
              gracePeriod: 5,
              successThreshold: 2,
              failureThreshold: 3,
              kind: "liveness",
            },
          ],
          files: [
            {
              guestPath: "/crowdsec-config/config.yaml",
              rawValue: crowdSecConfigOverride,
            },
            {
              guestPath: "/crowdsec-config/acquis.yaml",
              rawValue: crowdSecAcquisitionsConfig,
            },
          ],
          secrets: [
            {
              envVar: "BOUNCER_KEY_openresty",
            },
            {
              envVar: "ENROLL_KEY",
            },
          ],
        },
      ],
      mounts: [
        {
          path: "/var/lib/crowdsec/data/",
          sizeGb: 1,
          sizeGbLimit: crowdSecDataFork1.sizeGb,
          volume: crowdSecDataFork1.name,
        },
      ],
    },
    minSecretsVersion: secrets.version,
    queryParams: {
      delete: {
        // This is required since this machine uses a volume
        // and needs to support `deleteBeforeReplace`. And a
        // machine can only be deleted if it is stopped.
        // Setting `force` to `true` allows Fly to delete it
        // even if it's running.
        force: true,
      },
    },
  },
  {
    dependsOn: [crowdSecApp.app, secrets],
    replaceOnChanges: ["*"],
    deleteBeforeReplace: true,
  },
);

export const defaultCrowdSecUrl = `http://${crowdSecAppName}.flycast`;
