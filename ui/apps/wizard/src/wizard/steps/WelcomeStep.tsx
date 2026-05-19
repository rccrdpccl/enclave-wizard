import {
  Button,
  Content,
  Divider,
  Flex,
  FlexItem,
  Title,
} from "@patternfly/react-core";
import { css } from "@emotion/css";
import React from "react";
import type { ReactNode } from "react";
import { RedHatLogo } from "../../common/components/RedHatLogo.tsx";
import {
  BootstrapGlyph,
  DisconnectedGlyph,
  ShieldGlyph,
} from "../components/HighlightIcons.tsx";
import { useWizard } from "../WizardContext.tsx";

interface Highlight {
  icon: ReactNode;
  title: string;
  description: string;
}

const highlights: Highlight[] = [
  {
    icon: <BootstrapGlyph style={{ width: "2.25rem", height: "2.25rem" }} />,
    title: "Automated Bootstrap",
    description:
      "Image distribution and infrastructure setup, fully automated.",
  },
  {
    icon: (
      <DisconnectedGlyph style={{ width: "2.25rem", height: "2.25rem" }} />
    ),
    title: "Fully Disconnected",
    description:
      "Air-gapped by design for sovereign and secure environments.",
  },
  {
    icon: <ShieldGlyph style={{ width: "2.25rem", height: "2.25rem" }} />,
    title: "Smart Defaults",
    description:
      "Preconfigured settings for cluster management, GPU, and storage.",
  },
];

const styles = {
  container: css`
    text-align: center;
    padding: 2.75rem 1.75rem 2rem;
  `,
  heroTitle: css`
    line-height: 1.22;
    letter-spacing: -0.02em;
    font-weight: 500;
    font-size: clamp(1.5rem, 0.35rem + 2.1vw, 2.1rem);
  `,
  subtitle: css`
    color: var(--pf-t--global--color--subtle, #6a6e73);
    font-size: 1.0625rem;
    line-height: 1.6;
    max-width: 600px;
    margin: 0 auto;
  `,
  highlightList: css`
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: flex-start;
    gap: 0;
    margin-top: 2rem;

    @media (max-width: 48rem) {
      flex-direction: column;
      align-items: center;
    }
  `,
  highlightItem: css`
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    flex: 1;
    min-width: 0;
    max-width: 14rem;
    padding: 0 1rem;
  `,
  highlightIcon: css`
    color: var(--pf-t--global--icon--color--brand--default, #0066cc);
    margin-bottom: 0.75rem;
    line-height: 0;
  `,
  highlightTitle: css`
    color: #151515;
    margin-bottom: 0.25rem;
  `,
  highlightDescription: css`
    color: var(--pf-t--global--color--subtle, #6a6e73);
    font-size: 0.9rem;
    line-height: 1.5;
  `,
  dividerWrap: css`
    display: flex;
    align-items: center;
    padding: 0 0.5rem;
    align-self: stretch;

    @media (max-width: 48rem) {
      padding: 0.75rem 0;
      align-self: center;
      width: 60%;
    }
  `,
  cta: css`
    margin-top: 2.5rem;
  `,
};

export const WelcomeStep: React.FC = () => {
  const { dispatch } = useWizard();

  return (
    <Flex
      direction={{ default: "column" }}
      alignItems={{ default: "alignItemsCenter" }}
      gap={{ default: "gapLg" }}
      className={styles.container}
    >
      <FlexItem>
        <RedHatLogo />
      </FlexItem>
      <FlexItem>
        <Title headingLevel="h1" size="2xl" className={styles.heroTitle}>
          Welcome to Open Sovereign AI Cloud
        </Title>
      </FlexItem>
      <FlexItem>
        <Content component="p" className={styles.subtitle}>
          Deploy a sovereign, fully disconnected OpenShift environment with a
          simple, no-headache wizard.
        </Content>
      </FlexItem>
      <FlexItem style={{ width: "100%" }}>
        <div className={styles.highlightList} role="list">
          {highlights.map((h, i) => (
            <React.Fragment key={h.title}>
              {i > 0 && (
                <div className={styles.dividerWrap}>
                  <Divider
                    component="div"
                    orientation={{
                      default: "horizontal",
                      md: "vertical",
                    }}
                  />
                </div>
              )}
              <div className={styles.highlightItem} role="listitem">
                <div className={styles.highlightIcon} aria-hidden>
                  {h.icon}
                </div>
                <Title
                  headingLevel="h3"
                  size="md"
                  className={styles.highlightTitle}
                >
                  {h.title}
                </Title>
                <Content component="p" className={styles.highlightDescription}>
                  {h.description}
                </Content>
              </div>
            </React.Fragment>
          ))}
        </div>
      </FlexItem>
      <FlexItem className={styles.cta}>
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
