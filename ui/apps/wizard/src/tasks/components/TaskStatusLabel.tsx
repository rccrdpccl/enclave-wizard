import { Label, Spinner } from "@patternfly/react-core";
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  BanIcon,
} from "@patternfly/react-icons";
import type React from "react";

interface TaskStatusLabelProps {
  status: string;
}

export const TaskStatusLabel: React.FC<TaskStatusLabelProps> = ({ status }) => {
  switch (status) {
    case "running":
      return (
        <Label color="blue" icon={<Spinner size="sm" />}>
          Running
        </Label>
      );
    case "successful":
      return (
        <Label color="green" icon={<CheckCircleIcon />}>
          Successful
        </Label>
      );
    case "failed":
      return (
        <Label color="red" icon={<ExclamationCircleIcon />}>
          Failed
        </Label>
      );
    case "canceled":
      return (
        <Label color="grey" icon={<BanIcon />}>
          Canceled
        </Label>
      );
    default:
      return <Label>{status}</Label>;
  }
};
