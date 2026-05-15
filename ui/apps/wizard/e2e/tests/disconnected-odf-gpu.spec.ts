import { test, expect } from "@playwright/test";
import { WizardPage } from "../helpers/wizard-page";
import { WizardApi } from "../helpers/wizard-api";

test.describe("Disconnected + ODF + GPU scenario", () => {
  let wizard: WizardPage;
  let api: WizardApi;

  test.beforeEach(async ({ page, request, baseURL }) => {
    wizard = new WizardPage(page);
    api = new WizardApi(request, baseURL!);
  });

  test("completes full wizard flow with disconnected, ODF, and GPU enabled", async ({
    page,
  }) => {
    // a. Navigate and click Get started
    await wizard.goto();
    await wizard.clickGetStarted();

    // b. Select GPU & AI Workloads flavor and proceed
    await wizard.selectFlavor("GPU & AI Workloads");
    await wizard.clickNext();

    // c. Fill Landing Zone for disconnected mode
    await wizard.fillLandingZone({
      disconnected: true,
      lzBmcIP: "172.20.0.1",
      quayUser: "registry-admin",
      quayPassword: "odf-gpu-secret",
      quayBackend: "LocalStorage",
    });

    // d. Proceed to Hub Cluster
    await wizard.clickNext();

    // e. Fill Hub Cluster details with ODF and 3 hosts
    await wizard.fillHubCluster({
      baseDomain: "odf-gpu.enclave.io",
      clusterName: "gpu-mgmt",
      machineNetwork: "172.20.0.0/24",
      apiVIP: "172.20.0.200",
      ingressVIP: "172.20.0.201",
      rendezvousIP: "172.20.0.10",
      defaultDNS: "172.20.0.1",
      defaultGateway: "172.20.0.1",
      defaultPrefix: 24,
      enableOdf: true,
      odfExternalConfig: '{"clusterID":"e2e-ceph"}',
      pullSecret: '{"auths":{}}',
      sshPubKey: "ssh-rsa AAAA-gpu-test",
      hosts: [
        {
          name: "gpu-node-01",
          macAddress: "AA:BB:CC:DD:01:01",
          ipAddress: "172.20.0.11",
          redfish: "https://172.20.0.1:8443/redfish/v1/Systems/1",
          redfishUser: "admin",
          redfishPassword: "redfish01",
          rootDisk: "/dev/sda",
        },
        {
          name: "gpu-node-02",
          macAddress: "AA:BB:CC:DD:01:02",
          ipAddress: "172.20.0.12",
          redfish: "https://172.20.0.1:8443/redfish/v1/Systems/2",
          redfishUser: "admin",
          redfishPassword: "redfish02",
          rootDisk: "/dev/sda",
        },
        {
          name: "gpu-node-03",
          macAddress: "AA:BB:CC:DD:01:03",
          ipAddress: "172.20.0.13",
          redfish: "https://172.20.0.1:8443/redfish/v1/Systems/3",
          redfishUser: "admin",
          redfishPassword: "redfish03",
          rootDisk: "/dev/sda",
        },
      ],
    });

    // f. Click Next — should land on GPU & AI step (flavor-specific)
    await wizard.clickNext();

    // g. Verify nvidia-gpu checkbox is auto-selected by the GPU flavor
    const nvidiaCheckbox = page.locator("#gpu-ai-nvidia-gpu");
    await expect(nvidiaCheckbox).toBeChecked();

    // h. Click Next to Review
    await wizard.clickNext();

    // i. Verify global.yaml tab contains expected values
    const globalYaml = await wizard.getYamlContent("global.yaml");
    expect(globalYaml).toContain("odf-gpu.enclave.io");
    expect(globalYaml).toContain("nvidia-gpu");

    // j. Click Validate
    await wizard.clickValidate();

    // k. Click Next to Generate
    await wizard.clickNext();

    // l. Write configuration and wait for success
    await wizard.clickWriteConfiguration();
    await wizard.waitForWriteSuccess();

    // m. Verify via API that the written config includes odf, nvidia-gpu, and quayUser
    const config = await api.getConfig();
    const configStr = JSON.stringify(config);
    expect(configStr).toContain("odf");
    expect(configStr).toContain("nvidia-gpu");
    expect(configStr).toContain("registry-admin");
  });
});
