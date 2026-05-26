import type React from "react";
import { Navigate } from "react-router-dom";

export const ConditionalRedirect: React.FC = () => {
  return <Navigate to="/wizard" replace />;
};
