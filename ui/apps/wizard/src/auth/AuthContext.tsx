import React, {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";

interface AuthState {
  token: string | null;
  isAuthenticated: boolean;
  mustChangePassword: boolean;
}

interface AuthContextValue extends AuthState {
  login: (password: string) => Promise<void>;
  logout: () => void;
  changePassword: (
    currentPassword: string,
    newPassword: string,
  ) => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function getApiBasePath(): string {
  return window.location.origin;
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [state, setState] = useState<AuthState>({
    token: null,
    isAuthenticated: false,
    mustChangePassword: false,
  });

  const login = useCallback(async (password: string) => {
    const res = await fetch(`${getApiBasePath()}/api/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: "admin", password }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(
        (body as Record<string, string>).message || "Invalid credentials",
      );
    }

    const data = (await res.json()) as {
      token: string;
      mustChangePassword: boolean;
    };
    setState({
      token: data.token,
      isAuthenticated: true,
      mustChangePassword: data.mustChangePassword,
    });
  }, []);

  const logout = useCallback(() => {
    setState({
      token: null,
      isAuthenticated: false,
      mustChangePassword: false,
    });
  }, []);

  const changePassword = useCallback(
    async (currentPassword: string, newPassword: string) => {
      if (!state.token) {
        throw new Error("Not authenticated");
      }

      const res = await fetch(`${getApiBasePath()}/api/v1/auth/password`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${state.token}`,
        },
        body: JSON.stringify({ currentPassword, newPassword }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(
          (body as Record<string, string>).message ||
            "Failed to change password",
        );
      }

      const data = (await res.json()) as { token: string };
      setState({
        token: data.token,
        isAuthenticated: true,
        mustChangePassword: false,
      });
    },
    [state.token],
  );

  const value = useMemo<AuthContextValue>(
    () => ({
      ...state,
      login,
      logout,
      changePassword,
    }),
    [state, login, logout, changePassword],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
