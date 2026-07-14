import { useCallback, useState } from "react";
import { useAuth } from "../auth/AuthContext.tsx";

export interface UseFileUploadReturn {
  upload: (file: File, dest: string) => Promise<{ path: string }>;
  uploading: boolean;
  error: string | null;
  clearError: () => void;
}

export function useFileUpload(): UseFileUploadReturn {
  const { token } = useAuth();
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const upload = useCallback(
    async (file: File, dest: string): Promise<{ path: string }> => {
      setError(null);
      setUploading(true);
      try {
        const formData = new FormData();
        formData.append("file", file);
        formData.append("dest", dest);
        const headers: Record<string, string> = {};
        if (token) {
          headers.Authorization = `Bearer ${token}`;
        }
        const resp = await fetch("/api/v1/files", {
          method: "POST",
          headers,
          body: formData,
        });
        if (!resp.ok) {
          throw new Error(await resp.text());
        }
        return (await resp.json()) as { path: string };
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Upload failed";
        setError(msg);
        throw err;
      } finally {
        setUploading(false);
      }
    },
    [token],
  );

  const clearError = useCallback(() => setError(null), []);

  return { upload, uploading, error, clearError };
}
