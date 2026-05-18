import type { Download, Page } from "@playwright/test";
import { readFile } from "node:fs/promises";

export interface LandingZoneConfig {
  disconnected: boolean;
  lzBmcIP: string;
  lzBmcHostname?: string;
  quayUser?: string;
  quayPassword?: string;
  quayBackend?: string;
  rgw?: {
    access_key: string;
    secret_key: string;
    bucket_name: string;
    hostname: string;
  };
}

export interface HostEntry {
  name: string;
  macAddress: string;
  ipAddress: string;
  redfish: string;
  redfishUser: string;
  redfishPassword: string;
  rootDisk: string;
}

export interface HubClusterConfig {
  baseDomain: string;
  clusterName: string;
  machineNetwork: string;
  apiVIP: string;
  ingressVIP: string;
  rendezvousIP: string;
  defaultDNS: string;
  defaultGateway: string;
  defaultPrefix: number;
  pullSecret: string;
  sshPubKey: string;
  enableOdf?: boolean;
  odfExternalConfig?: string;
  hosts: HostEntry[];
}

export class WizardPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/wizard");
    await this.page.waitForLoadState("networkidle");
  }

  // --- Step: Welcome ---

  async clickGetStarted() {
    await this.page.getByRole("button", { name: "Get started" }).click();
  }

  // --- Step: Select Flavor ---

  async selectFlavor(title: string) {
    await this.page.getByText(title, { exact: false }).click();
  }

  async clickNext() {
    await this.page.getByRole("button", { name: "Next" }).click();
  }

  async clickBack() {
    await this.page.getByRole("button", { name: "Back" }).click();
  }

  // --- Step: Landing Zone ---

  async fillLandingZone(config: LandingZoneConfig) {
    await this.fillSchemaField("global.lzBmcIP", config.lzBmcIP);
    if (config.lzBmcHostname) {
      await this.fillSchemaField("global.lzBmcHostname", config.lzBmcHostname);
    }

    const checkbox = this.page.locator("#disconnected-toggle");
    const isChecked = await checkbox.isChecked();
    if (config.disconnected !== isChecked) {
      await checkbox.click();
    }

    if (config.disconnected) {
      if (config.quayUser) {
        await this.fillSchemaField("global.quayUser", config.quayUser);
      }
      if (config.quayPassword) {
        await this.fillSchemaField("global.quayPassword", config.quayPassword);
      }
      if (config.quayBackend) {
        await this.selectSchemaField("global.quayBackend", config.quayBackend);
      }
      if (config.rgw) {
        await this.page.fill("#rgw-access_key", config.rgw.access_key);
        await this.page.fill("#rgw-secret_key", config.rgw.secret_key);
        await this.page.fill("#rgw-bucket_name", config.rgw.bucket_name);
        await this.page.fill("#rgw-hostname", config.rgw.hostname);
      }
    }
  }

  // --- Step: Hub Cluster ---

  async fillHubCluster(config: HubClusterConfig) {
    await this.fillSchemaField("global.baseDomain", config.baseDomain);
    await this.fillSchemaField("global.clusterName", config.clusterName);
    await this.fillSchemaField("global.machineNetwork", config.machineNetwork);
    await this.fillSchemaField("global.apiVIP", config.apiVIP);
    await this.fillSchemaField("global.ingressVIP", config.ingressVIP);
    await this.fillSchemaField("global.rendezvousIP", config.rendezvousIP);
    await this.fillSchemaField("global.defaultDNS", config.defaultDNS);
    await this.fillSchemaField("global.defaultGateway", config.defaultGateway);
    await this.page.fill(
      "#field-global\\.defaultPrefix",
      String(config.defaultPrefix),
    );

    if (config.enableOdf) {
      const odfCheckbox = this.page.locator("#storage-odf");
      if (!(await odfCheckbox.isChecked())) {
        await odfCheckbox.click();
      }
      if (config.odfExternalConfig) {
        await this.page.fill("#odf-external-config", config.odfExternalConfig);
      }
    }

    await this.page.fill("#pull-secret", config.pullSecret);
    await this.page.fill("#ssh-pub-key", config.sshPubKey);

    for (const host of config.hosts) {
      await this.page.getByRole("button", { name: "Add node" }).click();
    }

    for (let i = 0; i < config.hosts.length; i++) {
      const host = config.hosts[i];
      await this.page.fill(`#node-${i}-name`, host.name);
      await this.page.fill(`#node-${i}-mac`, host.macAddress);
      await this.page.fill(`#node-${i}-ip`, host.ipAddress);
      await this.page.fill(`#node-${i}-redfish`, host.redfish);
      await this.page.fill(`#node-${i}-rfuser`, host.redfishUser);
      await this.page.fill(`#node-${i}-rfpass`, host.redfishPassword);
      await this.page.fill(`#node-${i}-rootdisk`, host.rootDisk);
    }
  }

  // --- Step: GPU & AI ---

  async selectGpuPlugin(pluginId: string) {
    const checkbox = this.page.locator(`#gpu-ai-${pluginId}`);
    if (!(await checkbox.isChecked())) {
      await checkbox.click();
    }
  }

  // --- Step: Review ---

  async getYamlContent(tab: string): Promise<string> {
    await this.page.getByRole("tab", { name: tab }).click();
    const editor = this.page.locator(".cm-content");
    return editor.textContent() ?? "";
  }

  async clickValidate() {
    await this.page.getByRole("button", { name: "Validate" }).click();
    await this.page.waitForSelector('[class*="alert"]', { timeout: 10_000 });
  }

  async clickCopyAll() {
    await this.page.getByRole("button", { name: "Copy all" }).click();
  }

  async clickDownloadFiles() {
    await this.page.getByRole("button", { name: "Download files" }).click();
  }

  async downloadConfigFiles(): Promise<
    Map<string, { download: Download; content: string }>
  > {
    const results = new Map<
      string,
      { download: Download; content: string }
    >();
    const downloads: Download[] = [];

    const collectDownloads = (d: Download) => downloads.push(d);
    this.page.on("download", collectDownloads);

    await this.page.getByRole("button", { name: "Download files" }).click();
    await this.page.waitForTimeout(1_000);

    this.page.removeListener("download", collectDownloads);

    for (const download of downloads) {
      const path = await download.path();
      if (path) {
        const content = await readFile(path, "utf-8");
        results.set(download.suggestedFilename(), { download, content });
      }
    }
    return results;
  }

  async isValidationSuccess(): Promise<boolean> {
    return this.page
      .locator('[class*="pf-m-success"]')
      .isVisible({ timeout: 5_000 });
  }

  // --- Step: Generate ---

  async clickWriteConfiguration() {
    await this.page
      .getByRole("button", { name: "Write configuration" })
      .click();
  }

  async waitForWriteSuccess() {
    await this.page.waitForSelector("text=Configuration written successfully", {
      timeout: 30_000,
    });
  }

  // --- Step: Provision (future) ---
  // These methods target UI elements that don't exist yet.
  // Update selectors when the provision UI is built.

  async clickProvision() {
    // TODO: Update selector when provision button is added to the UI
    await this.page
      .getByRole("button", { name: "Provision" })
      .click();
  }

  async getProvisionStatusFromPage(): Promise<string> {
    // TODO: Update selector when provision status UI is built
    // For now, returns empty string
    const statusEl = this.page.locator("[data-testid='provision-status']");
    if (await statusEl.isVisible({ timeout: 2_000 }).catch(() => false)) {
      return (await statusEl.textContent()) ?? "";
    }
    return "";
  }

  // --- Helpers ---

  private async fillSchemaField(path: string, value: string) {
    const selector = `#field-${path.replace(/\./g, "\\.")}`;
    await this.page.fill(selector, value);
  }

  private async selectSchemaField(path: string, value: string) {
    const selector = `#field-${path.replace(/\./g, "\\.")}`;
    await this.page.selectOption(selector, value);
  }
}
