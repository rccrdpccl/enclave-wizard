import { css } from "@emotion/css";

export const wizardStyles = {
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

  cardBody: css`
    padding: 2rem;
  `,

  footer: css`
    background-color: #fff;
    border-top: 1px solid #d2d2d2;
    flex-shrink: 0;
  `,

  footerInner: css`
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem 2rem;
  `,

  errorAlert: css`
    margin-bottom: 1rem;
  `,
};
