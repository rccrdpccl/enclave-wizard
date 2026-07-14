import { Button, Flex, FlexItem } from "@patternfly/react-core";
import type React from "react";
import { wizardStyles as styles } from "../wizardStyles.ts";

export const WizardFooter: React.FC<{
  onBack: () => void;
  onNext: () => void;
  isFirst: boolean;
  isLast: boolean;
  isContinue: boolean;
}> = ({ onBack, onNext, isFirst, isLast, isContinue }) => {
  return (
    <div className={styles.footer}>
      <div className={styles.footerInner}>
        <Flex justifyContent={{ default: "justifyContentSpaceBetween" }}>
          <FlexItem>
            <Button variant="secondary" onClick={onBack} isDisabled={isFirst}>
              Back
            </Button>
          </FlexItem>
          <FlexItem>
            <Button variant="primary" onClick={onNext} isDisabled={isLast}>
              {isContinue ? "Continue" : "Next"}
            </Button>
          </FlexItem>
        </Flex>
      </div>
    </div>
  );
};
