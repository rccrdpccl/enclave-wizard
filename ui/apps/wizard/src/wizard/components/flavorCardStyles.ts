import { css } from "@emotion/css";

export const flavorCardStyles = {
  card: (isSelected: boolean) => css`
    border: 2px solid ${isSelected ? "var(--pf-t--global--color--brand--default)" : "#d2d2d2"};
    cursor: pointer;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
    box-shadow: ${isSelected ? "0 0 0 1px var(--pf-t--global--color--brand--default)" : "none"};
  `,

  body: css`
    padding: 1.5rem;
  `,

  icon: (isSelected: boolean) => css`
    font-size: 2.5rem;
    color: ${isSelected ? "var(--pf-t--global--color--brand--default)" : "#6a6e73"};
  `,

  description: css`
    color: #6a6e73;
  `,
};
