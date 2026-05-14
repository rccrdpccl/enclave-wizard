import {
  Card,
  CardBody,
  Content,
  Flex,
  FlexItem,
  Label,
  Title,
} from "@patternfly/react-core";
import { CheckCircleIcon } from "@patternfly/react-icons";
import type React from "react";
import type { ReactNode } from "react";
import { flavorCardStyles as styles } from "./flavorCardStyles.ts";

interface FlavorCardProps {
  title: string;
  description: string;
  icon?: ReactNode;
  isSelected: boolean;
  onSelect: () => void;
}

export const FlavorCard: React.FC<FlavorCardProps> = ({
  title,
  description,
  icon,
  isSelected,
  onSelect,
}) => {
  return (
    // biome-ignore lint/a11y/useSemanticElements: PatternFly Card with isSelectable behaves as a button
    <Card
      isRounded
      isSelectable
      isSelected={isSelected}
      role="button"
      aria-pressed={isSelected}
      tabIndex={0}
      onClick={onSelect}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect();
        }
      }}
      className={styles.card(isSelected)}
    >
      <CardBody className={styles.body}>
        <Flex direction={{ default: "column" }} gap={{ default: "gapMd" }}>
          <Flex
            justifyContent={{ default: "justifyContentSpaceBetween" }}
            alignItems={{ default: "alignItemsCenter" }}
          >
            <FlexItem>
              {icon && <span className={styles.icon(isSelected)}>{icon}</span>}
            </FlexItem>
            <FlexItem>
              {isSelected ? (
                <Label color="blue" icon={<CheckCircleIcon />}>Selected</Label>
              ) : (
                <Label color="grey" variant="outline">Click to select</Label>
              )}
            </FlexItem>
          </Flex>
          <FlexItem>
            <Title headingLevel="h3" size="xl">{title}</Title>
          </FlexItem>
          <FlexItem>
            <Content component="p" className={styles.description}>{description}</Content>
          </FlexItem>
        </Flex>
      </CardBody>
    </Card>
  );
};
