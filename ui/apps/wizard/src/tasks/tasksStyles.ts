import { css } from "@emotion/css";

export const tasksStyles = {
  root: css`
    display: flex;
    flex-direction: column;
    height: 100vh;
    background-color: #f0f0f0;
  `,

  header: css`
    background-color: #fff;
    flex-shrink: 0;
  `,

  headerInner: css`
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem 2rem;
  `,

  content: css`
    flex: 1;
    overflow: auto;
    padding: 1.5rem 2rem;
  `,

  contentInner: css`
    max-width: 1200px;
    margin: 0 auto;
  `,

  logsContainer: css`
    background-color: #1e1e1e;
    color: #d4d4d4;
    font-family: "Red Hat Mono", "Liberation Mono", monospace;
    font-size: 0.8125rem;
    line-height: 1.5;
    padding: 1rem;
    border-radius: 4px;
    overflow: auto;
    max-height: 60vh;
    white-space: pre-wrap;
    word-break: break-all;
  `,

  navButton: css`
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 1rem;
    border: 1px solid var(--pf-t--global--border--color--default);
    border-radius: var(--pf-t--global--border--radius--small);
    color: var(--pf-t--global--text--color--regular);
    text-decoration: none;
    font-size: 0.875rem;
    &:hover {
      background-color: var(--pf-t--global--background--color--secondary--hover);
    }
  `,

  clickableRow: css`
    cursor: pointer;
    &:hover {
      background-color: #f0f0f0;
    }
  `,
};
