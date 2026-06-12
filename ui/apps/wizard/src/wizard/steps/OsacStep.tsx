import {
  Content,
  Flex,
  FlexItem,
  FormGroup,
  Radio,
  TextInput,
  Title,
} from "@patternfly/react-core";
import type React from "react";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles } from "./stepStyles.ts";

export const OsacStep: React.FC = () => {
  const { state, dispatch } = useWizard();
  const globalData = ((state.configData as Record<string, unknown>).global ??
    {}) as Record<string, unknown>;

  const osacProfile = (globalData.osacProfile as string) ?? "development";
  const aapLicenseFile = (globalData.osacAapLicenseFile as string) ?? "";
  const byoDatabase = (globalData.osacBYODatabase as boolean) ?? false;
  const databaseUrl = (globalData.osacDatabaseUrl as string) ?? "";

  const setField = (field: string, value: unknown) =>
    dispatch({ type: "SET_FIELD", path: `global.${field}`, value });

  return (
    <Flex direction={{ default: "column" }} gap={{ default: "gapLg" }}>
      <FlexItem>
        <Title headingLevel="h3" size="lg">
          OSAC Configuration
        </Title>
        <Content component="p" className={stepStyles.subtitle}>
          Configure the Open Sovereign AI Cloud platform settings.
        </Content>
      </FlexItem>

      <FlexItem>
        <FormGroup label="Deployment profile" isRequired fieldId="osac-profile">
          <Flex direction={{ default: "column" }} gap={{ default: "gapSm" }}>
            <Radio
              id="profile-development"
              name="osac-profile"
              label="Development (CaaS + VMaaS)"
              description="Full stack with cluster and VM management"
              isChecked={osacProfile === "development"}
              onChange={() => setField("osacProfile", "development")}
              isDisabled
            />
            <Radio
              id="profile-caas"
              name="osac-profile"
              label="CaaS"
              description="Cluster ordering and lifecycle management"
              isChecked={osacProfile === "caas"}
              onChange={() => setField("osacProfile", "caas")}
              isDisabled
            />
            <Radio
              id="profile-vmaas"
              name="osac-profile"
              label="VMaaS"
              description="Virtual machine provisioning and management"
              isChecked={osacProfile === "vmaas"}
              onChange={() => setField("osacProfile", "vmaas")}
              isDisabled
            />
          </Flex>
          <Content component="small" className={stepStyles.subtitle}>
            Profile is set automatically based on selected services.
          </Content>
        </FormGroup>
      </FlexItem>

      <FlexItem>
        <FormGroup
          label="AAP subscription file"
          isRequired
          fieldId="aap-license"
          helperText="Path to the AAP license manifest.zip on the landing zone. Obtain from access.redhat.com/management/subscription_allocations"
        >
          <TextInput
            id="aap-license"
            value={aapLicenseFile}
            onChange={(_e, val) => setField("osacAapLicenseFile", val)}
            placeholder="/path/to/aap-license.zip"
            validated={
              state.showValidation && !aapLicenseFile.trim()
                ? "error"
                : "default"
            }
          />
        </FormGroup>
      </FlexItem>

      <FlexItem>
        <FormGroup
          label="Database"
          fieldId="byo-database"
        >
          <Flex direction={{ default: "column" }} gap={{ default: "gapSm" }}>
            <Radio
              id="db-builtin"
              name="byo-database"
              label="Built-in PostgreSQL"
              description="Deploy a managed PostgreSQL instance (recommended for dev/test)"
              isChecked={!byoDatabase}
              onChange={() => setField("osacBYODatabase", false)}
            />
            <Radio
              id="db-external"
              name="byo-database"
              label="Bring your own database"
              description="Connect to an existing PostgreSQL instance"
              isChecked={byoDatabase}
              onChange={() => setField("osacBYODatabase", true)}
            />
          </Flex>
        </FormGroup>
      </FlexItem>

      {byoDatabase && (
        <FlexItem>
          <FormGroup
            label="Database URL"
            isRequired
            fieldId="database-url"
            helperText="PostgreSQL connection string"
          >
            <TextInput
              id="database-url"
              value={databaseUrl}
              onChange={(_e, val) => setField("osacDatabaseUrl", val)}
              placeholder="postgres://user@host:5432/dbname?sslmode=require"
              validated={
                state.showValidation && byoDatabase && !databaseUrl.trim()
                  ? "error"
                  : "default"
              }
            />
          </FormGroup>
        </FlexItem>
      )}
    </Flex>
  );
};
