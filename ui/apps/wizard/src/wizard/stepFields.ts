export const STEP_REQUIRED_FIELDS: Record<string, string[]> = {
  "landing-zone": [
    "global.lzBmcIP",
    "global.quayUser",
    "global.quayPassword",
    "global.quayBackend",
  ],
  "hub-cluster": [
    "global.baseDomain",
    "global.clusterName",
    "global.machineNetwork",
    "global.apiVIP",
    "global.ingressVIP",
    "global.rendezvousIP",
    "global.defaultDNS",
    "global.defaultGateway",
    "global.defaultPrefix",
    "global.blockStorageBackend",
    "global.pullSecret",
    "global.sshPubPath",
    "global.agent_hosts",
  ],
};
