import * as pulumi from "@pulumi/pulumi";

const config = new pulumi.Config();

// https://docs.fly.io/reference/regions#regions
export const region = "ord";

export const customDomain = "mydomain.com";

export const flyOrgSlug = config.require("flyOrgSlug");

export const crowdSecEnrollKey = config.requireSecret("crowdSecEnrollKey");
