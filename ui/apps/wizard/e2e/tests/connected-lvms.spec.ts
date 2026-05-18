import { test, expect } from "@playwright/test";
import { WizardPage } from "../helpers/wizard-page";
import { WizardApi } from "../helpers/wizard-api";

test.describe("Connected + LVMS scenario", () => {
  let wizard: WizardPage;
  let api: WizardApi;

  test.beforeEach(async ({ page, request, baseURL }) => {
    wizard = new WizardPage(page);
    api = new WizardApi(request, baseURL!);
  });

  test("completes the wizard end-to-end with connected LVMS config", async () => {
    // a. Navigate to wizard and click Get started
    await wizard.goto();
    await wizard.clickGetStarted();

    // b. Skip flavor selection (no CaaS or GPU needed) — just click Next
    await wizard.clickNext();

    // c. Fill Landing Zone: connected mode with BMC IP
    await wizard.fillLandingZone({
      disconnected: false,
      lzBmcIP: "10.10.50.1",
    });

    // d. Click Next to proceed to Hub Cluster
    await wizard.clickNext();

    // e. Fill Hub Cluster configuration
    await wizard.fillHubCluster({
      baseDomain: "connected-lvms.lab.local",
      clusterName: "edge-conn",
      machineNetwork: "10.10.50.0/24",
      apiVIP: "10.10.50.200",
      ingressVIP: "10.10.50.201",
      rendezvousIP: "10.10.50.10",
      defaultDNS: "10.10.50.1",
      defaultGateway: "10.10.50.1",
      defaultPrefix: 24,
      pullSecret: '{"auths":{}}',
      sshPubKey: "ssh-rsa AAAA-test-key",
      hosts: [
        {
          name: "ctrl-01",
          macAddress: "00:00:00:00:01:01",
          ipAddress: "10.10.50.11",
          redfish: "10.10.50.1",
          redfishUser: "admin",
          redfishPassword: "password",
          rootDisk: "/dev/sda",
        },
        {
          name: "ctrl-02",
          macAddress: "00:00:00:00:01:02",
          ipAddress: "10.10.50.12",
          redfish: "10.10.50.1",
          redfishUser: "admin",
          redfishPassword: "password",
          rootDisk: "/dev/sda",
        },
        {
          name: "ctrl-03",
          macAddress: "00:00:00:00:01:03",
          ipAddress: "10.10.50.13",
          redfish: "10.10.50.1",
          redfishUser: "admin",
          redfishPassword: "password",
          rootDisk: "/dev/sda",
        },
      ],
    });

    // f. Click Next to proceed to Review step
    await wizard.clickNext();

    // g. Verify YAML content in global.yaml tab
    const yamlContent = await wizard.getYamlContent("global.yaml");
    expect(yamlContent).toContain("connected-lvms.lab.local");
    expect(yamlContent).toContain("edge-conn");

    // h. Click Validate and verify validation succeeds
    await wizard.clickValidate();
    const isValid = await wizard.isValidationSuccess();
    expect(isValid).toBe(true);

    // i. Click Next to proceed to Generate step
    await wizard.clickNext();

    // j. Click "Write configuration" and wait for success
    await wizard.clickWriteConfiguration();
    await wizard.waitForWriteSuccess();

    // k. Verify via API: read config back and check key fields match
    const config = await api.getConfig();
    const global = config.global as Record<string, unknown>;
    expect(global.baseDomain).toBe("connected-lvms.lab.local");
    expect(global.clusterName).toBe("edge-conn");
    expect(global.machineNetwork).toBe("10.10.50.0/24");
    expect(global.apiVIP).toBe("10.10.50.200");
    expect(global.ingressVIP).toBe("10.10.50.201");
    expect(global.rendezvousIP).toBe("10.10.50.10");
    expect(global.lzBmcIP).toBe("10.10.50.1");
  });
});
