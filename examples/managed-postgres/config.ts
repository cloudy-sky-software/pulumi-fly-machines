import * as pulumi from "@pulumi/pulumi";

const config = new pulumi.Config();
export const orgSlug = config.require("orgSlug");

export const appName = "pulumi-fly-machines-test-app";
