import {
  Card,
  CardBody,
  FormGroup,
  TextInput,
  Title,
} from "@patternfly/react-core";
import type React from "react";
import { hostEntryCardStyles as styles } from "./hostEntryCardStyles.ts";

interface HostEntry {
  name: string;
  macAddress: string;
  ipAddress: string;
  redfish: string;
  redfishUser: string;
  redfishPassword: string;
  rootDisk: string;
}

interface HostEntryCardProps {
  index: number;
  host: HostEntry;
  onChange: (host: HostEntry) => void;
  label?: string;
}

export type { HostEntry };

export const HostEntryCard: React.FC<HostEntryCardProps> = ({
  index,
  host,
  onChange,
  label = "Host",
}) => {
  const prefix = `${label.toLowerCase().replace(/\s+/g, "-")}-${index}`;

  const update = (field: keyof HostEntry, value: string) =>
    onChange({ ...host, [field]: value });

  return (
    <Card isRounded isCompact>
      <CardBody>
        <Title headingLevel="h4" size="md">
          {label} {index + 1}
        </Title>
        <div className={styles.grid}>
          <FormGroup label="Name" isRequired fieldId={`${prefix}-name`}>
            <TextInput
              id={`${prefix}-name`}
              value={host.name}
              onChange={(_e, v) => update("name", v)}
              isRequired
            />
          </FormGroup>
          <FormGroup label="MAC address" isRequired fieldId={`${prefix}-mac`}>
            <TextInput
              id={`${prefix}-mac`}
              value={host.macAddress}
              onChange={(_e, v) => update("macAddress", v)}
              isRequired
            />
          </FormGroup>
          <FormGroup label="IP address" isRequired fieldId={`${prefix}-ip`}>
            <TextInput
              id={`${prefix}-ip`}
              value={host.ipAddress}
              onChange={(_e, v) => update("ipAddress", v)}
              isRequired
            />
          </FormGroup>
          <FormGroup
            label="Redfish IP"
            isRequired
            fieldId={`${prefix}-redfish`}
          >
            <TextInput
              id={`${prefix}-redfish`}
              value={host.redfish}
              onChange={(_e, v) => update("redfish", v)}
              isRequired
            />
          </FormGroup>
          <FormGroup
            label="Redfish user"
            isRequired
            fieldId={`${prefix}-rfuser`}
          >
            <TextInput
              id={`${prefix}-rfuser`}
              value={host.redfishUser}
              onChange={(_e, v) => update("redfishUser", v)}
              isRequired
            />
          </FormGroup>
          <FormGroup
            label="Redfish password"
            isRequired
            fieldId={`${prefix}-rfpass`}
          >
            <TextInput
              id={`${prefix}-rfpass`}
              type="password"
              value={host.redfishPassword}
              onChange={(_e, v) => update("redfishPassword", v)}
              isRequired
            />
          </FormGroup>
          <div className={styles.fullWidth}>
          <FormGroup
            label="Root Disk Path"
            isRequired
            fieldId={`${prefix}-rootdisk`}
          >
            <TextInput
              id={`${prefix}-rootdisk`}
              value={host.rootDisk}
              onChange={(_e, v) => update("rootDisk", v)}
              isRequired
            />
          </FormGroup>
          </div>
        </div>
      </CardBody>
    </Card>
  );
};
