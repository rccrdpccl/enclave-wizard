import { describe, expect, it } from "vitest";

// The validation logic is extracted as a pure function for testing purposes.
// We test the validation rules directly rather than through the hook,
// since the hook just wraps useCallback around the same logic.

import {
  validateFields,
  validateHostEntries,
} from "../../schema/schemaUtils.ts";
import { STEP_REQUIRED_FIELDS } from "../stepFields.ts";

// Minimal schema structure that validates required fields
const MOCK_SCHEMA = {
  properties: {
    global: {
      type: "object",
      properties: {
        lzBmcIP: { type: "string", doc: "Landing Zone BMC IP" },
        baseDomain: { type: "string", doc: "Base Domain" },
        clusterName: { type: "string", doc: "Cluster Name" },
        machineNetwork: { type: "string", doc: "Machine Network" },
        apiVIP: { type: "string", doc: "API VIP" },
        ingressVIP: { type: "string", doc: "Ingress VIP" },
        rendezvousIP: { type: "string", doc: "Rendezvous IP" },
        defaultDNS: { type: "string", doc: "DNS Server" },
        defaultGateway: { type: "string", doc: "Default Gateway" },
        defaultPrefix: { type: "number", doc: "Subnet Prefix" },
        storage_plugin: { type: "string", doc: "Storage Plugin" },
        pullSecret: { type: "object", doc: "Pull Secret" },
        sshPubKey: { type: "string", doc: "SSH Public Key" },
        agent_hosts: {
          type: "array",
          items: {
            type: "object",
            properties: {
              name: { type: "string" },
              macAddress: { type: "string" },
              ipAddress: { type: "string" },
              redfish: { type: "string" },
              redfishUser: { type: "string" },
              redfishPassword: { type: "string" },
              rootDisk: { type: "string" },
            },
            required: [
              "name",
              "macAddress",
              "ipAddress",
              "redfish",
              "redfishUser",
              "redfishPassword",
              "rootDisk",
            ],
          },
        },
      },
      required: [
        "lzBmcIP",
        "baseDomain",
        "clusterName",
        "machineNetwork",
        "apiVIP",
        "ingressVIP",
        "rendezvousIP",
        "defaultDNS",
        "defaultGateway",
        "defaultPrefix",
        "storage_plugin",
        "pullSecret",
        "sshPubKey",
      ],
    },
  },
};

describe("step validation rules", () => {
  describe("landing-zone required fields", () => {
    it("returns errors for missing required fields", () => {
      const fields = STEP_REQUIRED_FIELDS["landing-zone"];
      const errors = validateFields(MOCK_SCHEMA, fields, { global: {} });
      expect(errors.length).toBeGreaterThan(0);
      expect(errors[0].path).toBe("global.lzBmcIP");
    });

    it("returns no errors when required fields are filled", () => {
      const fields = STEP_REQUIRED_FIELDS["landing-zone"];
      const errors = validateFields(MOCK_SCHEMA, fields, {
        global: { lzBmcIP: "192.168.1.1" },
      });
      expect(errors).toHaveLength(0);
    });
  });

  describe("storage required fields", () => {
    it("returns error for missing storage_plugin", () => {
      const fields = STEP_REQUIRED_FIELDS.storage;
      const errors = validateFields(MOCK_SCHEMA, fields, { global: {} });
      expect(errors.some((e) => e.path === "global.storage_plugin")).toBe(true);
    });
  });

  describe("hub-cluster required fields", () => {
    it("validates multiple required hub-cluster fields", () => {
      const fields = STEP_REQUIRED_FIELDS["hub-cluster"].filter(
        (f) => f !== "global.agentHosts",
      );
      const errors = validateFields(MOCK_SCHEMA, fields, { global: {} });
      // Should have errors for baseDomain, clusterName, machineNetwork, etc.
      expect(errors.length).toBeGreaterThanOrEqual(5);
    });

    it("returns no errors when all hub-cluster fields provided", () => {
      const fields = STEP_REQUIRED_FIELDS["hub-cluster"].filter(
        (f) => f !== "global.agentHosts",
      );
      const errors = validateFields(MOCK_SCHEMA, fields, {
        global: {
          baseDomain: "test.local",
          clusterName: "hub",
          machineNetwork: "10.0.0.0/24",
          apiVIP: "10.0.0.100",
          ingressVIP: "10.0.0.101",
          rendezvousIP: "10.0.0.10",
          defaultDNS: "10.0.0.1",
          defaultGateway: "10.0.0.1",
          defaultPrefix: 24,
          pullSecret: { auths: {} },
          sshPubKey: "ssh-rsa AAAA",
        },
      });
      expect(errors).toHaveLength(0);
    });
  });

  describe("host validation", () => {
    it("validates required host entry fields", () => {
      const hosts = [{ name: "", macAddress: "", ipAddress: "" }];
      const errors = validateHostEntries(MOCK_SCHEMA, hosts, "Node");
      // Missing required fields should produce errors
      expect(errors.length).toBeGreaterThan(0);
    });

    it("passes with valid host entries", () => {
      const hosts = [
        {
          name: "cp-0",
          macAddress: "00:00:00:00:00:01",
          ipAddress: "10.0.0.10",
          redfish: "10.0.0.1:8000",
          redfishUser: "admin",
          redfishPassword: "pass",
          rootDisk: "/dev/sda",
        },
      ];
      const errors = validateHostEntries(MOCK_SCHEMA, hosts, "Node");
      expect(errors).toHaveLength(0);
    });
  });
});
