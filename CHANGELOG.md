## 0.0.14

Updated API spec to the latest which includes
detailed descriptions for many properties that
carry over as doc comments for all languages.

## 0.0.13

Fix for some enums not generating correctly.

Some enum types that were previously typed
as `any` will now be properly typed.

There are no changes to resources.

## 0.0.12

- Adds `listPostgresPlans` function to list the available
  Postgres plans.
- You can now specify the `plan` when forking an existing
  Postgres cluster.

## 0.0.11

Handle `Secret` and `Secretkey` resources for pseudo-id
assignment since the API does not assign one.

## 0.0.10

Fixes the `headers` type in `FlyHTTPResponseOptions`.

## 0.0.9

Fixes the schema type for some properties that have
an `additionalProperties` configuration.

## 0.0.8

Update the dotnet SDK root namespace to `CloudySkySoftware.Pulumi` to
avoid collision with the reserved namespace `Pulumi`.

## 0.0.7

Update `pulumi-provider-framework` to the latest version.

## 0.0.6

Set NODE_AUTH_TOKEN for the SDK build step in `release.yml`.

## 0.0.5

Add a separate job to publish the release in `release.yml`.

## 0.0.4

Change goreleaser config to create draft GH releases in order
to allow parallel jobs to upload artifacts to the same release.
Publish the release using GH CLI which will make the release
immutable.

## 0.0.3

Fix release workflow issues.

## 0.0.2

So close!

## 0.0.1

First release! Maybe this will work?
