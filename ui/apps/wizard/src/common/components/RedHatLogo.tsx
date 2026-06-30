import { css } from "@emotion/css";
import { CloudIcon } from "@patternfly/react-icons";
import type React from "react";

const brandStyle = css`
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: inherit;
`;

const iconStyle = css`
  font-size: 2rem;
  color: var(--pf-t--global--icon--color--brand--default);
`;

const textStyle = css`
  display: flex;
  flex-direction: column;
  line-height: 1.1;
`;

const titleStyle = css`
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: 0.02em;
`;

const subtitleStyle = css`
  font-size: 0.6875rem;
  color: var(--pf-t--global--text--color--subtle);
  letter-spacing: 0.04em;
  text-transform: uppercase;
`;

export const RedHatLogo: React.FC<{ width?: number }> = () => {
  return (
    <div className={brandStyle}>
      <CloudIcon className={iconStyle} />
      <div className={textStyle}>
        <span className={titleStyle}>OSAC</span>
        <span className={subtitleStyle}>Open Sovereign AI Cloud</span>
      </div>
    </div>
  );
};
