import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

import { customDomain, flyOrgSlug, region } from "./config.js";

import { readFileSync } from "fs";
import path from "path";
import { cwd } from "process";

import { defaultInternalFlyAppUrl, internalAppMachine } from "./private-app.js";
import { FlyApp } from "./flyApp.js";
import {
  crowdSecMachine,
  defaultCrowdSecUrl,
  openRestyApiKey,
} from "./crowdsec.js";

// This is the port number used in nginx.conf.
const reverseProxyPort = 9000;
const reverseProxyAppName = "reverse-proxy";

const reverseProxyApp = new FlyApp(
  reverseProxyAppName,
  {
    customDomain,
    ipAssignments: ["shared_v4"],
    orgSlug: flyOrgSlug,
  },
  { dependsOn: internalAppMachine },
);

const openrestyDefaultNginxConfig = Buffer.from(
  readFileSync(
    path.join(
      cwd().endsWith("infra") ? "." : path.join("src", "infra"),
      "configs",
      "nginx",
      "default.conf",
    ),
    {
      encoding: "utf-8",
    },
  )
    .replaceAll("${{REVERSE_PROXY_PORT}}", `${reverseProxyPort}`)
    .replaceAll("${{INTERNAL_APP_URL}}", `${defaultInternalFlyAppUrl}/`),
  "utf-8",
).toString("base64");

const openrestyMainNginxConfig = Buffer.from(
  readFileSync(
    path.join(
      cwd().endsWith("infra") ? "." : path.join("src", "infra"),
      "configs",
      "nginx",
      "nginx.conf",
    ),
    {
      encoding: "utf-8",
    },
  ),
  "utf-8",
).toString("base64");

const banHtml = Buffer.from(
  readFileSync(
    path.join(
      cwd().endsWith("infra") ? "." : path.join("src", "infra"),
      "configs",
      "nginx",
      "ban.html",
    ),
    {
      encoding: "utf-8",
    },
  ),
  "utf-8",
).toString("base64");

const crowdSecOpenRestyConfig = Buffer.from(
  readFileSync(
    path.join(
      cwd().endsWith("infra") ? "." : path.join("src", "infra"),
      "configs",
      "nginx",
      "crowdsec_openresty.conf",
    ),
    {
      encoding: "utf-8",
    },
  ),
  "utf-8",
).toString("base64");

const secrets = new fly.apps.v1.Secrets("openresty-secrets", {
  appName: reverseProxyAppName,
  values: {
    API_KEY: openRestyApiKey,
  },
});

const openrestyMachine = new fly.apps.v1.Machine(
  "openresty-machine",
  {
    appName: reverseProxyAppName,
    region,
    config: {
      guest: {
        cpuKind: "shared",
        cpus: 1,
        memoryMb: 256,
      },
      init: {},
      cacheDisk: {
        path: "/tmp/openresty",
        sizeMb: 100,
      },
      restart: {
        maxRetries: 10,
        policy: "on-failure",
      },
      metadata: {
        fly_platform_version: "v2",
        fly_process_group: "openresty",
      },
      services: [
        {
          ports: [{ port: 443, handlers: ["tls", "http"] }],
          protocol: "tcp",
          internalPort: reverseProxyPort,
          autostop: "suspend",
          autostart: true,
          minMachinesRunning: 1,
        },
      ],
      containers: [
        {
          name: "openresty",
          // https://hub.docker.com/r/crowdsecurity/openresty
          image: "crowdsecurity/openresty:v1.2.5",
          env: {
            API_URL: defaultCrowdSecUrl,
          },
          files: [
            {
              guestPath: "/etc/nginx/conf.d/default.conf",
              rawValue: openrestyDefaultNginxConfig,
            },
            {
              guestPath: "/usr/local/openresty/nginx/conf/nginx.conf",
              rawValue: openrestyMainNginxConfig,
            },
            {
              guestPath: "/var/lib/crowdsec/lua/templates/ban.html",
              rawValue: banHtml,
            },
            {
              guestPath: "/etc/nginx/conf.d/crowdsec_openresty.conf",
              rawValue: crowdSecOpenRestyConfig,
            },
          ],
          secrets: [
            {
              envVar: "API_KEY",
            },
          ],
        },
      ],
    },
    minSecretsVersion: secrets.version,
  },
  { dependsOn: [reverseProxyApp, internalAppMachine, crowdSecMachine] },
);

// Change the suffix arg to re-trigger verification.
reverseProxyApp.verifyDomain(customDomain, "1");

export const defaultReverseProxyFlyAppUrl = `https://${reverseProxyAppName}.fly.dev`;
export const reverseProxyCertVerification =
  reverseProxyApp.cert?.dnsRequirements;
