import type React from "react";

export const RedHatLogo: React.FC<{ width?: number }> = ({ width = 200 }) => {
  return (
    <img
      src="/redhat-osac.png"
      alt="Red Hat Open Sovereign AI Cloud"
      width={width}
      style={{ maxWidth: "100%", height: "auto" }}
    />
  );
};
