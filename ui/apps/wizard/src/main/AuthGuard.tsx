import React from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../auth/AuthContext.tsx";

export const AuthGuard: React.FC = () => {
  const { isAuthenticated, mustChangePassword } = useAuth();

  if (!isAuthenticated || mustChangePassword) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
};
