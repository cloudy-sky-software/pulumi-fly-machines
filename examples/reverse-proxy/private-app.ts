import * as pulumi from "@pulumi/pulumi";
import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

import {
  azureClientId,
  azureClientSecret,
  azureSubscriptionId,
  azureTenantId,
  flyOrgSlug,
  keygenServerImageTag,
  keyName,
  pocketBaseImageTag,
  publicKeyStr,
  region,
  smtpPassword,
  vaultName,
} from "./config.js";

export const internalAppName = "internal-app";
export const internalAppPort = 8080;

const internalApp = new fly.apps.v1.App(internalAppName, {
  name: internalAppName,
  orgSlug: flyOrgSlug,
});

const privateIp = new fly.apps.v1.AppIPAssignment("deskypus-pb-private_v6-ip", {
  appName: internalAppName,
  type: "private_v6",
});

export const internalAppMachine = new fly.apps.v1.Machine(
  "internal-app-machine",
  {
    appName: internalAppName,
    region,
    config: {
      guest: {
        cpuKind: "shared",
        cpus: 1,
        memoryMb: 512,
      },
      init: {},
      restart: {
        maxRetries: 10,
        policy: "on-failure",
      },
      metadata: {
        fly_platform_version: "v2",
        fly_process_group: "app",
      },
      services: [
        {
          ports: [{ port: 80, handlers: ["http"] }],
          protocol: "tcp",
          internalPort: internalAppPort,
          autostop: "suspend",
          autostart: true,
          minMachinesRunning: 1,
        },
      ],
      containers: [
        {
          name: "echo-server",
          image: "ealen/echo-server:latest",
          env: {
            // Disable some echo server features.
            // <echo-server>
            ENABLE__ENVIRONMENT: "false",
            LOGS__LEVEL: "info",
            // </echo-server>
            PORT: `${internalAppPort}`,
          },
          healthchecks: [
            {
              http: {
                port: internalAppPort,
                method: "GET",
                path: "/healthcheck",
                scheme: "http",
              },
              interval: 10,
              timeout: 5,
              gracePeriod: 5,
              successThreshold: 2,
              failureThreshold: 3,
            },
          ],
        },
      ],
    },
  },
  {
    dependsOn: [privateIp, internalApp],
  },
);

export const defaultInternalFlyAppUrl = `http://${internalAppName}.flycast`;
