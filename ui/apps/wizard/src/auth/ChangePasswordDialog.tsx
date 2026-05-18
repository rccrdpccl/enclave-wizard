import {
  Alert,
  Button,
  Content,
  Form,
  FormGroup,
  FormHelperText,
  HelperText,
  HelperTextItem,
  Modal,
  ModalBody,
  ModalFooter,
  ModalHeader,
  TextInput,
} from "@patternfly/react-core";
import React, { useState } from "react";
import { useAuth } from "./AuthContext.tsx";

const MIN_PASSWORD_LENGTH = 8;

interface ChangePasswordDialogProps {
  currentPassword: string;
}

export const ChangePasswordDialog: React.FC<ChangePasswordDialogProps> = ({
  currentPassword,
}) => {
  const { changePassword } = useAuth();

  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const tooShort = newPassword.length > 0 && newPassword.length < MIN_PASSWORD_LENGTH;
  const mismatch = confirmPassword.length > 0 && newPassword !== confirmPassword;
  const canSubmit =
    newPassword.length >= MIN_PASSWORD_LENGTH &&
    newPassword === confirmPassword &&
    !submitting;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;

    setError(null);
    setSubmitting(true);

    try {
      await changePassword(currentPassword, newPassword);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to change password");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen
      variant="small"
      aria-label="Change your password"
      disableFocusTrap={false}
    >
      <ModalHeader title="Change your password" />
      <ModalBody>
        <Content component="p" style={{ marginBottom: "1.5rem" }}>
          Your initial password must be changed before continuing.
        </Content>

        {error && (
          <Alert
            variant="danger"
            title={error}
            isInline
            style={{ marginBottom: "1rem" }}
          />
        )}

        <Form id="change-password-form" onSubmit={handleSubmit}>
          <FormGroup
            label="New password"
            fieldId="new-password"
            isRequired
          >
            <TextInput
              id="new-password"
              type="password"
              value={newPassword}
              onChange={(_e, val) => setNewPassword(val)}
              validated={tooShort ? "error" : "default"}
              autoFocus
              isRequired
              aria-label="New password"
            />
            {tooShort && (
              <FormHelperText>
                <HelperText>
                  <HelperTextItem variant="error">
                    Password must be at least {MIN_PASSWORD_LENGTH} characters
                  </HelperTextItem>
                </HelperText>
              </FormHelperText>
            )}
          </FormGroup>

          <FormGroup
            label="Confirm password"
            fieldId="confirm-password"
            isRequired
          >
            <TextInput
              id="confirm-password"
              type="password"
              value={confirmPassword}
              onChange={(_e, val) => setConfirmPassword(val)}
              validated={mismatch ? "error" : "default"}
              isRequired
              aria-label="Confirm password"
            />
            {mismatch && (
              <FormHelperText>
                <HelperText>
                  <HelperTextItem variant="error">
                    Passwords do not match
                  </HelperTextItem>
                </HelperText>
              </FormHelperText>
            )}
          </FormGroup>
        </Form>
      </ModalBody>
      <ModalFooter>
        <Button
          variant="primary"
          type="submit"
          form="change-password-form"
          isLoading={submitting}
          isDisabled={!canSubmit}
        >
          Change password
        </Button>
      </ModalFooter>
    </Modal>
  );
};
