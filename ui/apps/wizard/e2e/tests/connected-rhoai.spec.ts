import { test, expect } from "@playwright/test";
import { WizardPage } from "../helpers/wizard-page";
import { WizardApi } from "../helpers/wizard-api";

test.describe("Connected + RHOAI scenario", () => {
  let wizard: WizardPage;
  let api: WizardApi;

  test.beforeEach(async ({ page, request, baseURL }) => {
    wizard = new WizardPage(page);
    api = new WizardApi(request, baseURL!);
  });

  test("configures a connected cluster with GPU & AI workloads and RHOAI plugin", async ({
    page,
  }) => {

    // a. Navigate and click Get started
    await wizard.goto();
    await wizard.clickGetStarted();

    // b. Select GPU & AI Workloads flavor, click Next
    await wizard.selectFlavor("GPU & AI Workloads");
    await wizard.clickNext();

    // c. Fill Landing Zone: connected mode
    await wizard.fillLandingZone({
      disconnected: false,
      lzBmcIP: "192.168.100.1",
    });

    // d. Click Next
    await wizard.clickNext();

    // e. Fill Hub Cluster
    await wizard.fillHubCluster({
      baseDomain: "rhoai.lab.example.com",
      clusterName: "ai-edge",
      machineNetwork: "192.168.100.0/24",
      apiVIP: "192.168.100.200",
      ingressVIP: "192.168.100.201",
      rendezvousIP: "192.168.100.10",
      defaultDNS: "192.168.100.1",
      defaultGateway: "192.168.100.1",
      defaultPrefix: 24,
      pullSecret: '{"auths":{}}',
      sshPubKey: "ssh-rsa AAAA-rhoai-test",
      hosts: [
        {
          name: "ai-node-01",
          macAddress: "00:00:00:00:01:01",
          ipAddress: "192.168.100.11",
          redfish: "192.168.100.1",
          redfishUser: "admin",
          redfishPassword: "password",
          rootDisk: "/dev/sda",
        },
        {
          name: "ai-node-02",
          macAddress: "00:00:00:00:01:02",
          ipAddress: "192.168.100.12",
          redfish: "192.168.100.1",
          redfishUser: "admin",
          redfishPassword: "password",
          rootDisk: "/dev/sda",
        },
        {
          name: "ai-node-03",
          macAddress: "00:00:00:00:01:03",
          ipAddress: "192.168.100.13",
          redfish: "192.168.100.1",
          redfishUser: "admin",
          redfishPassword: "password",
          rootDisk: "/dev/sda",
        },
      ],
    });

    // f. Click Next to GPU & AI step
    await wizard.clickNext();

    // g. Select openshift-ai plugin (should auto-select nvidia-gpu)
    await wizard.selectGpuPlugin("openshift-ai");

    // h. Verify nvidia-gpu is checked and disabled (auto-selected by RHOAI dependency)
    const nvidiaCheckbox = page.locator("#gpu-ai-nvidia-gpu");
    await expect(nvidiaCheckbox).toBeChecked();
    await expect(nvidiaCheckbox).toBeDisabled();

    // i. Click Next to Review
    await wizard.clickNext();

    // j. Verify global.yaml contains expected values
    const globalYaml = await wizard.getYamlContent("global.yaml");
    expect(globalYaml).toContain("rhoai.lab.example.com");
    expect(globalYaml).toContain("openshift-ai");
    expect(globalYaml).toContain("nvidia-gpu");

    // k. Click Download files to test the download flow
    await wizard.clickDownloadFiles();

    // l. Click Validate
    await wizard.clickValidate();
    const validationPassed = await wizard.isValidationSuccess();
    expect(validationPassed).toBe(true);

    // m. Click Next to Generate
    await wizard.clickNext();

    // n. Click "Write configuration", wait for success
    await wizard.clickWriteConfiguration();
    await wizard.waitForWriteSuccess();

    // o. Verify via API: config has openshift-ai, nvidia-gpu, no quayUser
    const config = await api.getConfig();
    const configStr = JSON.stringify(config);
    expect(configStr).toContain("openshift-ai");
    expect(configStr).toContain("nvidia-gpu");
    expect(configStr).not.toContain("quayUser");
  });
});
