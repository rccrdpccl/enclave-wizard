// NOTE: Provisioning API is not yet implemented. This test uses fake responses.
// When the real API lands, only WizardApi.triggerProvision and
// WizardApi.getProvisionStatus need to change — these tests stay the same.

import { test, expect } from "@playwright/test";
import { WizardPage } from "../helpers/wizard-page";
import { WizardApi } from "../helpers/wizard-api";

const minimalHubConfig = {
  baseDomain: "provision-test.local",
  clusterName: "prov-cl",
  machineNetwork: "10.0.0.0/24",
  apiVIP: "10.0.0.10",
  ingressVIP: "10.0.0.11",
  rendezvousIP: "10.0.0.12",
  defaultDNS: "10.0.0.1",
  defaultGateway: "10.0.0.1",
  defaultPrefix: 24,
  pullSecret: '{"auths":{}}',
  sshPubKey: "ssh-ed25519 AAAA fake-key",
  hosts: [
    {
      name: "node-0",
      macAddress: "AA:BB:CC:DD:EE:00",
      ipAddress: "10.0.0.20",
      redfish: "https://10.0.0.1/redfish/v1/Systems/1",
      redfishUser: "admin",
      redfishPassword: "password",
      rootDisk: "/dev/sda",
    },
    {
      name: "node-1",
      macAddress: "AA:BB:CC:DD:EE:01",
      ipAddress: "10.0.0.21",
      redfish: "https://10.0.0.1/redfish/v1/Systems/2",
      redfishUser: "admin",
      redfishPassword: "password",
      rootDisk: "/dev/sda",
    },
    {
      name: "node-2",
      macAddress: "AA:BB:CC:DD:EE:02",
      ipAddress: "10.0.0.22",
      redfish: "https://10.0.0.1/redfish/v1/Systems/3",
      redfishUser: "admin",
      redfishPassword: "password",
      rootDisk: "/dev/sda",
    },
  ],
};

test.describe("Provision flow", () => {
  let wizard: WizardPage;
  let api: WizardApi;

  test.beforeEach(async ({ page, request, baseURL }) => {
    wizard = new WizardPage(page);
    api = new WizardApi(request, baseURL!);
  });

  test("full wizard flow ending with provision trigger", async () => {

    // Navigate through the wizard: Welcome -> Flavor -> Landing Zone -> Hub Cluster -> Review -> Generate
    await wizard.goto();
    await wizard.clickGetStarted();

    // Skip flavor selection (use default)
    await wizard.clickNext();

    // Landing Zone — connected mode, minimal config
    await wizard.fillLandingZone({
      disconnected: false,
      lzBmcIP: "10.0.0.1",
    });
    await wizard.clickNext();

    // Hub Cluster — minimal 3-node config
    await wizard.fillHubCluster(minimalHubConfig);
    await wizard.clickNext();

    // Review step — proceed to Generate
    await wizard.clickNext();

    // Generate step — write configuration
    await wizard.clickWriteConfiguration();
    await wizard.waitForWriteSuccess();

    // Read the written config back via the API
    const config = await api.getConfig();
    expect(config).toBeDefined();

    // Trigger provisioning (currently returns fake "accepted" response)
    const provisionResponse = await api.triggerProvision(config);
    expect(provisionResponse.state).toBe("accepted");
    expect(provisionResponse.id).toBeTruthy();

    // Check provision status (currently returns fake "completed" response)
    const status = await api.getProvisionStatus(provisionResponse.id);
    expect(status.state).toBe("completed");
    expect(status.progress).toBe(100);
  });

  test("config can be downloaded from review step before provisioning", async () => {

    // Navigate to the Review step with minimal config
    await wizard.goto();
    await wizard.clickGetStarted();
    await wizard.clickNext(); // skip flavor

    await wizard.fillLandingZone({
      disconnected: false,
      lzBmcIP: "10.0.0.1",
    });
    await wizard.clickNext();

    await wizard.fillHubCluster(minimalHubConfig);
    await wizard.clickNext();

    // Now on the Review step — click Download files
    await wizard.clickDownloadFiles();

    // Read the YAML content from the global.yaml tab
    const yamlContent = await wizard.getYamlContent("global.yaml");
    expect(yamlContent).toContain("provision-test.local");
  });

  test("provision status polling", async () => {

    // Write config via API
    const config = {
      global: {
        baseDomain: "poll-test.local",
        clusterName: "poll-cl",
        lzBmcIP: "10.0.0.1",
      },
    };
    await api.writeConfig(config);

    // Trigger provision via API
    const provisionResponse = await api.triggerProvision(config);
    expect(provisionResponse.state).toBe("accepted");

    // Poll getProvisionStatus in a loop (max 5 iterations, 1s delay)
    // When the real API lands, increase timeout and delay to account for
    // actual provisioning duration (e.g., 30+ iterations with 10s delay).
    let finalStatus = await api.getProvisionStatus(provisionResponse.id);
    for (let i = 0; i < 5; i++) {
      finalStatus = await api.getProvisionStatus(provisionResponse.id);
      if (finalStatus.state === "completed" || finalStatus.state === "failed") {
        break;
      }
      await new Promise((resolve) => setTimeout(resolve, 1_000));
    }

    // Verify final state
    expect(finalStatus.state).toBe("completed");
    expect(finalStatus.progress).toBe(100);
  });
});
