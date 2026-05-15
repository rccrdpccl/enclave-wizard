export const STEP_REQUIRED_FIELDS: Record<string, string[]> = {
  "landing-zone": [
    "global.lzBmcIP",
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
    "global.pullSecret",
    "global.sshPubPath",
    "global.agent_hosts",
  ],
};
