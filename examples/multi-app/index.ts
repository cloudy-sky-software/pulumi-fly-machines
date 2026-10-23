import * as pulumi from "@pulumi/pulumi";
import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

import { customDomain, flyOrgSlug, mySecret } from "./config.js";

// https://docs.fly.io/reference/regions#regions
const region = "ord";

// The process group name can be anything.
const privateAppServiceProcessGroupName = "private";

const privateAppServicePort = 3000;
const publicAppServicePort = 8080;

const privateApp = new fly.apps.v1.App("private-app", {
    name: "private-app",
    orgSlug: flyOrgSlug,
});

// Example of how Fly App Secrets can also be in .env file formats
// and mounted as a file on the machine.
//
// Note: pulumi.interpolate is used here since `mySecret` is an Output
// type and cannot be interpolated using regular
const envFile = pulumi.interpolate`MY_SECRET=${mySecret}`;
const privateAppSecretEnvFile = new fly.apps.v1.Secret("private-app-secrets", {
    appName: privateApp.name,
    secretName: "private-app-secrets",
    value: pulumi.secret(
        envFile.apply((f) => Buffer.from(f, "utf-8").toString("base64")),
    ),
});

const privateAppMachine = new fly.apps.v1.Machine(
    "private-app-machine",
    {
        appName: privateApp.name,
        region,
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
                fly_process_group: privateAppServiceProcessGroupName,
            },
            containers: [
                {
                    name: "keygen",
                    image: "ealen/echo-server:latest",
                    env: {
                        // Bind to the `fly-local-6pn` host for Fly's
                        // 6PN private networking.
                        // This DOES NOT provider load balancing.
                        // If load balancing is required, allocate
                        // a `private_v6` IP address to the app that
                        // this machine belongs to and bind to `0.0.0.0`.
                        // Then the app and its machines are reachable
                        // via `<app-name>.flycast` domain.
                        // https://docs.fly.io/networking/flycast
                        BIND_TO_HOST: "fly-local-6pn",
                        // Disable some echo server's feature.
                        // <echo-server>
                        ENABLE__ENVIRONMENT: "false",
                        LOGS__LEVEL: "info",
                        // </echo-server>
                        PORT: `${privateAppServicePort}`,
                    },
                    files: [
                        {
                            guestPath: "/etc/secrets/.env",
                            secretName: privateAppSecretEnvFile.name,
                        },
                    ],
                },
            ],
        },
        minSecretsVersion: privateAppSecretEnvFile.version,
        queryParams: {
            delete: {
                // Passes the `force` query param for the delete
                // operation which will force the machine to be
                // deleted even if it is running.
                force: true,
            },
        },
    },
    { dependsOn: privateAppSecretEnvFile },
);

const publicApp = new fly.apps.v1.App(
    "public-app",
    {
        name: "public-app",
        orgSlug: flyOrgSlug,
    },
    { dependsOn: privateApp },
);

const ip = new fly.apps.v1.AppIPAssignment("app-ip", {
    appName: publicApp.name,
    type: "shared_v4",
});

const dataVolume = new fly.apps.v1.Volume("my-volume", {
    appName: publicApp.name,
    autoBackupEnabled: true,
    encrypted: true,
    sizeGb: 10,
    name: "data",
    region,
});

// Use the Secrets resource if you don't need to mount an .env-style
// file for your app. When the key name matches an env var defined
// in the secrets config for a machine running in this app, the
// secret value is automatically injected as that env var.
const publicAppSecrets = new fly.apps.v1.Secrets("public-app-secrets", {
    appName: publicApp.name,
    values: {
        MY_SECRET: mySecret,
    },
});

const publicAppMachine = new fly.apps.v1.Machine(
    "public-app-machine",
    {
        appName: publicApp.name,
        region,
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
                fly_process_group: "pocketbase",
            },
            services: [
                {
                    ports: [{ port: 443, handlers: ["tls", "http"] }],
                    protocol: "tcp",
                    internalPort: publicAppServicePort,
                    autostop: "suspend",
                    autostart: true,
                    minMachinesRunning: 1,
                },
            ],
            containers: [
                {
                    name: "pocketbase",
                    image: "registry.fly.io/deskypus-pb:sha-8bebb7c",
                    // image: "ealen/echo-server:latest",
                    env: {
                        // This env var can be used by this app to talk to
                        // the private app.
                        PRIVATE_APP_URL: pulumi.interpolate`http://${privateAppServiceProcessGroupName}.process.${privateApp.name}.internal:${privateAppServicePort}`,
                        // Disable some echo server's feature.
                        // <echo-server>
                        ENABLE__ENVIRONMENT: "false",
                        LOGS__LEVEL: "info",
                        // </echo-server>
                        PORT: `${publicAppServicePort}`,
                    },
                    healthchecks: [
                        {
                            http: {
                                port: publicAppServicePort,
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
                    secrets: [
                        {
                            // The name of this env var matches with the name
                            // of the secret in the `values` map of the
                            // `publicAppSecrets` app-level secrets.
                            envVar: "MY_SECRET",
                        },
                        {
                            envVar: "TIGRIS_STORAGE_SECRET_KEY",
                        },
                    ],
                },
            ],
            mounts: [
                {
                    name: "data",
                    path: "/data",
                    sizeGb: 1,
                    sizeGbLimit: dataVolume.sizeGb,
                    volume: dataVolume.name,
                },
            ],
        },
        minSecretsVersion: publicAppSecrets.version,
        queryParams: {
            delete: {
                // Passes the `force` query param for the delete
                // operation which will force the machine to be
                // deleted even if it is running.
                force: true,
            },
        },
    },
    {
        dependsOn: [ip, privateAppMachine, dataVolume, publicAppSecrets],
        // `replaceOnChanges` and `deleteBeforeReplace` are required because this
        // machine uses a volume and any changes that trigger a new machine
        // to be created require that the current machine be detached from
        // the volume.
        //
        // You are responsible for detaching volumes.
        //
        // Attaching a volume also prevents the machine from scaling up.
        replaceOnChanges: ["*"],
        deleteBeforeReplace: true,
    },
);

const cert = new fly.apps.v1.AppCertificatesAcme(
    "cert",
    {
        appName: publicApp.name,
        hostname: customDomain,
    },
    { dependsOn: publicAppMachine },
);

// Recreate this resource anytime you want to trigger
// another check. You can modify its name so that Pulumi
// would recreate it. Deleting this won't affect the
// actual certificate. This resource is a convience
// to trigger the check for validation records.
new fly.apps.v1.AppCertificatesCheck(
    "cert-check",
    {
        appName: publicApp.name,
        hostname: customDomain,
    },
    { dependsOn: [cert, publicAppMachine], deleteBeforeReplace: true },
);

export const defaultPublicAppUrl = pulumi.interpolate`https://${publicApp.name}.fly.dev`;
export const publicAppDomainOwnershipRecord = cert.dnsRequirements.apply(
    (r) => r?.ownership,
);
export const publicAppCnameRequirement = cert.dnsRequirements.apply(
    (r) => r?.cname,
);
