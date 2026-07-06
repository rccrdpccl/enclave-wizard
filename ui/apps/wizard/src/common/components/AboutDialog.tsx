import { css } from "@emotion/css";
import {
  Button,
  DescriptionList,
  DescriptionListDescription,
  DescriptionListGroup,
  DescriptionListTerm,
  Modal,
  ModalBody,
  ModalHeader,
} from "@patternfly/react-core";
import { OutlinedQuestionCircleIcon } from "@patternfly/react-icons";
import type React from "react";
import { useCallback, useEffect, useState } from "react";

const buttonStyle = css`
  && {
    color: var(--pf-t--global--text--color--subtle);
    font-size: 1rem;
    padding: 0.25rem 0.5rem;
  }
  &&:hover {
    color: var(--pf-t--global--text--color--regular);
  }
`;

interface VersionInfo {
  wizardVersion: string;
  enclaveVersion: string;
}

export const AboutDialog: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [versions, setVersions] = useState<VersionInfo | null>(null);

  const fetchVersions = useCallback(async () => {
    try {
      const resp = await fetch("/api/v1/version");
      if (resp.ok) {
        setVersions(await resp.json());
      }
    } catch {
      // version endpoint unavailable
    }
  }, []);

  useEffect(() => {
    if (isOpen && !versions) {
      fetchVersions();
    }
  }, [isOpen, versions, fetchVersions]);

  return (
    <>
      <Button
        variant="plain"
        aria-label="About"
        onClick={() => setIsOpen(true)}
        className={buttonStyle}
      >
        <OutlinedQuestionCircleIcon />
      </Button>
      <Modal
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        aria-label="About"
        variant="small"
      >
        <ModalHeader title="About" />
        <ModalBody>
          <DescriptionList isHorizontal>
            <DescriptionListGroup>
              <DescriptionListTerm>Enclave Wizard</DescriptionListTerm>
              <DescriptionListDescription>
                {versions?.wizardVersion ?? "—"}
              </DescriptionListDescription>
            </DescriptionListGroup>
            <DescriptionListGroup>
              <DescriptionListTerm>Enclave</DescriptionListTerm>
              <DescriptionListDescription>
                {versions?.enclaveVersion ?? "—"}
              </DescriptionListDescription>
            </DescriptionListGroup>
          </DescriptionList>
        </ModalBody>
      </Modal>
    </>
  );
};
