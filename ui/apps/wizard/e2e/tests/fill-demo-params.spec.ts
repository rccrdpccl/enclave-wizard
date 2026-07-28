import { test } from "@playwright/test";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { WizardApi } from "../helpers/wizard-api";
import { WizardPage } from "../helpers/wizard-page";

interface DemoParams {
  infra: {
    machineNetwork: string;
    gateway: string;
    apiVIP: string;
    ingressVIP: string;
    rendezvousIP: string;
    baseDomain: string;
    clusterName: string;
    defaultPrefix: number;
    bmc: { endpoint: string; user: string; password: string };
    hosts: Array<{
      name: string;
      mac: string;
      ip: string;
      disk: string;
      uuid: string;
    }>;
  };
  wizard: { url: string; password: string };
}

async function loadDemoParams(): Promise<DemoParams> {
  const paramsPath = resolve(__dirname, "../../../../..", "demo-params.json");
  return JSON.parse(await readFile(paramsPath, "utf-8"));
}

test("fill wizard from demo-params.json (stop before deploy)", async ({
  page,
  request,
  baseURL,
}) => {
  test.setTimeout(120_000);
  const params = await loadDemoParams();
  const password = process.env.WIZARD_PASSWORD ?? params.wizard.password;
  const bmcIP = params.infra.bmc.endpoint.replace(/:\d+$/, "");

  const newPassword = "pleaseletmein";

  // Write config via API — login, change password if needed, merge demo-params
  const api = new WizardApi(request, baseURL!);
  const loginResult = await api.login(password);
  if (loginResult.mustChangePassword) {
    await api.changePassword(password, newPassword);
  }

  const existing = (await api.getConfig()) as Record<
    string,
    Record<string, unknown>
  >;

  // Hub cluster fields
  existing.global.baseDomain = params.infra.baseDomain;
  existing.global.clusterName = params.infra.clusterName;
  existing.global.machineNetwork = params.infra.machineNetwork;
  existing.global.apiVIP = params.infra.apiVIP;
  existing.global.ingressVIP = params.infra.ingressVIP;
  existing.global.rendezvousIP = params.infra.rendezvousIP;
  existing.global.defaultDNS = params.infra.gateway;
  existing.global.defaultGateway = params.infra.gateway;
  existing.global.defaultPrefix = params.infra.defaultPrefix;
  existing.global.lzBmcIP = bmcIP;
  existing.global.disconnected = false;
  existing.global.storage_plugin = "lvms";
  existing.global.quayBackend = "LocalStorage";
  existing.global.quayUser = "admin";
  existing.global.quayPassword = "quaypass";
  existing.global.agent_hosts = params.infra.hosts.map((h) => ({
    name: h.name,
    macAddress: h.mac,
    ipAddress: h.ip,
    rootDisk: h.disk,
    redfish: bmcIP,
    redfishUser: params.infra.bmc.user,
    redfishPassword: params.infra.bmc.password,
  }));

  // Clear empty CaaS discovery hosts that would fail server-side validation
  if (existing.cloudInfra) {
    const hosts = existing.cloudInfra.discovery_hosts;
    if (Array.isArray(hosts)) {
      existing.cloudInfra.discovery_hosts = hosts.filter(
        (h: Record<string, string>) => h.name || h.ipAddress || h.macAddress,
      );
    }
  }

  await api.writeConfig(existing);

  // Open the wizard UI — config loads from what we just wrote
  await page.goto("/wizard?skip_validation");
  await page.waitForLoadState("networkidle");
  const uiPassword = loginResult.mustChangePassword ? newPassword : password;
  await page.fill("#login-password", uiPassword);
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForLoadState("networkidle");

  const wizard = new WizardPage(page);

  // Navigate to Review: Welcome → Select → Configure (sub-steps) → Review
  await wizard.clickGetStarted();
  await wizard.clickNext(); // Select → Configure

  let safetyCount = 0;
  while (safetyCount < 15) {
    const continueBtn = page.getByRole("button", { name: "Continue" });
    if (await continueBtn.isVisible({ timeout: 500 }).catch(() => false)) {
      await continueBtn.click();
      await page.waitForTimeout(300);
      safetyCount++;
    } else {
      break;
    }
  }
  await wizard.clickNext(); // Configure → Review

  // Pause so user can inspect and click deploy manually
  await page.pause();
});
