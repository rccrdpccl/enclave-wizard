import { Spinner } from "@patternfly/react-core";
import type React from "react";
import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { useTasksApi } from "../api/useTasksApi.ts";

export const ConditionalRedirect: React.FC = () => {
  const { listTasks } = useTasksApi();
  const [target, setTarget] = useState<string | null>(null);

  useEffect(() => {
    listTasks()
      .then((result) => {
        setTarget(result.runs.length > 0 ? "/tasks" : "/wizard");
      })
      .catch(() => {
        setTarget("/wizard");
      });
  }, [listTasks]);

  if (!target) {
    return <Spinner aria-label="Loading..." />;
  }

  return <Navigate to={target} replace />;
};
