import { css } from "@emotion/css";

export const stepStyles = {
  sectionTitle: css`
    margin-top: 1.5rem;
  `,

  firstSectionTitle: css`
    margin-top: 1rem;
  `,

  subtitle: css`
    margin-top: 0.5rem;
    color: #6a6e73;
  `,

  removeButton: css`
    margin-top: 2rem;
  `,

  emptyHint: css`
    color: var(--pf-t--global--color--subtle);
    margin-top: 0.5rem;
  `,

  warningHint: css`
    color: var(--pf-t--global--color--status--warning--default);
    margin-top: 0.5rem;
  `,

  hostSection: css`
    margin-top: 0.5rem;
  `,

  certsSection: css`
    margin-top: 1.5rem;
  `,

  welcomeContainer: css`
    text-align: center;
    padding: 3rem 1rem;
  `,

  welcomeDescription: css`
    max-width: 600px;
  `,

  validationError: css`
    color: var(--pf-t--global--color--status--danger--default);
  `,

  arrayItemRow: css`
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
  `,
};
