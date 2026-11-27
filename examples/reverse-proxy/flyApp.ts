import { ComponentResource, ComponentResourceOptions } from "@pulumi/pulumi";
import * as fly from "@cloudyskysoftware/pulumi-fly-machines";

interface FlyAppArgs {
  /**
   * Setting custom domain will only create a certificate
   * but will not trigger a certificate check.
   *
   * **IMPORTANT**: To complete domain ownership verification,
   * call `verifyDomain` once the necessary records have been
   * added to your hosted zone in the domain registrar.
   */
  customDomain?: string;
  ipAssignments?: fly.apps.v1.AppIPAssignmentType[];
  orgSlug: string;
}

/**
 * A component resource to create commonly used resources
 * together when creating a new Fly App.
 */
export class FlyApp extends ComponentResource {
  private readonly ipAssignments?: fly.apps.v1.AppIPAssignment[];
  private readonly name: string;

  public readonly app: fly.apps.v1.App;
  public readonly defaultAppUrl: string;
  public readonly cert?: fly.apps.v1.AppCertificatesAcme;

  constructor(name: string, args: FlyAppArgs, opts?: ComponentResourceOptions) {
    super("deskypus:flyApp:FlyApp", name, args, opts);

    this.name = name;

    this.app = new fly.apps.v1.App(
      name,
      {
        name,
        orgSlug: args.orgSlug,
      },
      { parent: this },
    );

    if (args.ipAssignments) {
      this.ipAssignments = args.ipAssignments.map(
        (ipType) =>
          new fly.apps.v1.AppIPAssignment(
            `${name}-${ipType}-ip`,
            {
              appName: name,
              type: ipType,
            },
            { parent: this.app },
          ),
      );
    }

    this.defaultAppUrl = `https://${name}.fly.dev`;

    if (args.customDomain) {
      this.cert = new fly.apps.v1.AppCertificatesAcme(
        `${name}-${args.customDomain.replaceAll(".", "-")}-cert`,
        {
          appName: name,
          hostname: args.customDomain,
        },
        { parent: this.app },
      );
    }

    super.registerOutputs({});
  }

  verifyDomain(customDomain: string, suffix: string) {
    if (!this.cert) {
      throw new Error(
        "App does not have a certificate. Set a custom domain first.",
      );
    }

    new fly.apps.v1.AppCertificatesCheck(
      `${this.name}-cert-check-${suffix}`,
      {
        appName: this.name,
        hostname: customDomain,
      },
      { parent: this.cert },
    );
  }
}
