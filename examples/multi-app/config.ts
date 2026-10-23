import * as pulumi from "@pulumi/pulumi";

const config = new pulumi.Config();

export const customDomain = "mydomain.com";
export const flyOrgSlug = config.require("flyOrgSlug");

export const mySecret = config.requireSecret("mySecret");
