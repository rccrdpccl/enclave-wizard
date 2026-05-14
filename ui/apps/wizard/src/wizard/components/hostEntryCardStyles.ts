import { css } from "@emotion/css";

export const hostEntryCardStyles = {
  grid: css`
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
    margin-top: 0.5rem;
  `,

  fullWidth: css`
    grid-column: 1 / -1;
  `,
};
