import { Button, Content, Flex, FlexItem, Title } from "@patternfly/react-core";
import type React from "react";
import { RedHatLogo } from "../../common/components/RedHatLogo.tsx";
import { useWizard } from "../WizardContext.tsx";
import { stepStyles as styles } from "./stepStyles.ts";

export const WelcomeStep: React.FC = () => {
  const { dispatch } = useWizard();

  return (
    <Flex
      direction={{ default: "column" }}
      alignItems={{ default: "alignItemsCenter" }}
      gap={{ default: "gapLg" }}
      className={styles.welcomeContainer}
    >
      <FlexItem>
        <RedHatLogo />
      </FlexItem>
      <FlexItem>
        <Title headingLevel="h1" size="2xl">
          Open Sovereign AI Cloud
        </Title>
      </FlexItem>
      <FlexItem>
        <Content component="p" className={styles.welcomeDescription}>
          Deploy a sovereign, fully disconnected OpenShift environment with a
          simple, no-headache wizard.
        </Content>
      </FlexItem>
      <FlexItem>
        <Button
          variant="primary"
          size="lg"
          onClick={() => dispatch({ type: "SET_STEP", step: 1 })}
        >
          Get started
        </Button>
      </FlexItem>
    </Flex>
  );
};
