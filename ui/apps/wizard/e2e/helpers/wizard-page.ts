import type { Download, Page } from "@playwright/test";
import { readFile } from "node:fs/promises";

export interface LandingZoneConfig {
  disconnected: boolean;
  lzBmcIP: string;
  lzBmcHostname?: string;
}

export interface StorageConfig {
  storagePlugin?: string;
  quayUser?: string;
  quayPassword?: string;
  quayBackend?: string;
  rgw?: {
    access_key: string;
    secret_key: string;
    bucket_name: string;
    hostname: string;
  };
  enableOdf?: boolean;
  odfExternalConfig?: string;
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
  hosts: HostEntry[];
}

export interface OsacConfig {
  aapLicenseFilename?: string;
  byoDatabase?: boolean;
  databaseUrl?: string;
  rhbkInstances?: number;
  rhbkDeployDatabase?: boolean;
  rhbkDbSize?: string;
}

export class WizardPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/wizard");
    await this.page.waitForLoadState("networkidle");
  }

  // --- Auth: Login & Change Password ---

  async login(password: string) {
    await this.page.fill("#login-password", password);
    await this.page.getByRole("button", { name: "Sign in" }).click();
  }

  async changePassword(newPassword: string) {
    await this.page.fill("#new-password", newPassword);
    await this.page.fill("#confirm-password", newPassword);
    await this.page.getByRole("button", { name: "Change password" }).click();
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
    const next = this.page.getByRole("button", { name: "Next" });
    const cont = this.page.getByRole("button", { name: "Continue" });
    if (await cont.isVisible({ timeout: 1000 }).catch(() => false)) {
      await cont.click();
    } else {
      await next.click();
    }
    await this.page.waitForTimeout(500);
  }

  async navigateToSubStep(label: string) {
    await this.page.getByText(label, { exact: true }).click();
    await this.page.waitForTimeout(500);
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
  }

  // --- Step: Storage ---

  async fillStorage(config: StorageConfig) {
    if (config.storagePlugin === "odf") {
      await this.page.locator("#storage-odf").click();
      if (config.odfExternalConfig) {
        await this.page.fill("#odf-external-config", config.odfExternalConfig);
      }
    } else if (config.storagePlugin === "vast-csi") {
      await this.page.locator("#storage-vast").click();
    } else {
      await this.page.locator("#storage-lvms").click();
    }

    if (config.quayBackend === "RadosGWStorage") {
      await this.page.locator("#quay-rgw").click();
    } else {
      await this.page.locator("#quay-local").click();
    }

    if (config.rgw) {
      await this.page.fill("#rgw-access_key", config.rgw.access_key);
      await this.page.fill("#rgw-secret_key", config.rgw.secret_key);
      await this.page.fill("#rgw-bucket_name", config.rgw.bucket_name);
      await this.page.fill("#rgw-hostname", config.rgw.hostname);
    }

    if (config.quayUser) {
      await this.page.fill("#quay-user", config.quayUser);
    }
    if (config.quayPassword) {
      await this.page.fill("#quay-password", config.quayPassword);
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

    await this.page.fill("#pull-secret", config.pullSecret);
    await this.page.fill("#ssh-pub-key", config.sshPubKey);

    // Remove any existing hosts first
    while (await this.page.locator('button[aria-label^="Remove node"]').first().isVisible({ timeout: 500 }).catch(() => false)) {
      await this.page.locator('button[aria-label^="Remove node"]').first().click();
      await this.page.waitForTimeout(200);
    }

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

  // --- Step: GPU / VMaaS ---

  async enableGpuPassthrough() {
    const checkbox = this.page.locator("#enable-gpu");
    if (!(await checkbox.isChecked())) {
      await checkbox.click();
    }
  }

  // --- Step: OSAC Platform ---

  async fillOsac(config: OsacConfig) {
    // Upload AAP license manifest
    const filename = config.aapLicenseFilename ?? "manifest.zip";
    const fileInput = this.page.locator("input[type='file']").first();
    await fileInput.setInputFiles({
      name: filename,
      mimeType: "application/zip",
      buffer: Buffer.from("PK\x03\x04fake-manifest-zip-content"),
    });
    await this.page.getByText("Saved to:").waitFor({ timeout: 10_000 });

    // Database backend
    if (config.byoDatabase) {
      await this.page.locator("#db-external").click();
      if (config.databaseUrl) {
        await this.page.fill("#database-url", config.databaseUrl);
      }
    }

    // RHBK instances (NumberInput — target the inner input element)
    if (config.rhbkInstances != null) {
      const input = this.page.locator("#rhbk-instances input[type='number']");
      await input.fill(String(config.rhbkInstances));
    }

    // RHBK deploy database toggle
    if (config.rhbkDeployDatabase === false) {
      const toggle = this.page.locator("#rhbk-deploy-database");
      if (await toggle.isChecked()) {
        await toggle.click();
      }
    }

    // RHBK database size
    if (config.rhbkDbSize) {
      await this.page.fill("#rhbk-db-size", config.rhbkDbSize);
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
    await this.page.waitForSelector('[class*="alert"]', { timeout: 30_000 });
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
      .getByText("Configuration is valid")
      .isVisible({ timeout: 5_000 });
  }

  // --- Step: Generate ---

  async clickWriteConfiguration() {
    // Legacy: tries "Write configuration" first, falls back to "Deploy"
    const writeBtn = this.page.getByRole("button", { name: "Write configuration" });
    const deployBtn = this.page.getByRole("button", { name: "Deploy" });
    if (await writeBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await writeBtn.click();
    } else {
      await deployBtn.click();
    }
  }

  async waitForWriteSuccess() {
    // Accept either "Configuration written" or deploy task started
    const written = this.page.getByText("Configuration written successfully");
    const taskStarted = this.page.getByText("Deploy");
    await Promise.race([
      written.waitFor({ timeout: 30_000 }).catch(() => {}),
      this.page.waitForTimeout(5_000),
    ]);
  }

  // --- Step: Provision ---

  async clickProvision() {
    await this.page
      .getByRole("button", { name: "Provision" })
      .click();
  }

  async getProvisionStatusFromPage(): Promise<string> {
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
