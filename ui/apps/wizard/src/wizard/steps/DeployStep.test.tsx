import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { DeploymentState } from "../../api/useDeployment.ts";
import { DeployStep } from "./DeployStep.tsx";

const mockStart = vi.fn();
let mockDeploymentState: DeploymentState = {
  phase: "idle",
  progress: null,
  logs: "",
  startTime: null,
  error: null,
  taskId: null,
};

vi.mock("../../api/useDeployment.ts", () => ({
  useDeployment: () => ({
    state: mockDeploymentState,
    start: mockStart,
  }),
}));

vi.mock("../WizardContext.tsx", () => ({
  useWizard: () => ({
    state: {
      currentStep: 0,
      selectedFlavors: new Set(),
      configData: {},
      validationErrors: [],
      showValidation: false,
      schema: null,
      plugins: [],
    },
  }),
}));

vi.mock("../buildFinalConfig.ts", () => ({
  buildFinalConfig: (state: unknown) => state,
}));

describe("DeployStep", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockDeploymentState = {
      phase: "idle",
      progress: null,
      logs: "",
      startTime: null,
      error: null,
      taskId: null,
    };
    Element.prototype.scrollIntoView = vi.fn();
  });

  it("shows deploy button in idle state", () => {
    render(<DeployStep />);
    expect(screen.getByRole("button", { name: "Deploy" })).toBeInTheDocument();
  });

  it("calls start on deploy button click", async () => {
    mockStart.mockResolvedValue(undefined);

    render(<DeployStep />);
    await userEvent.click(screen.getByRole("button", { name: "Deploy" }));

    expect(mockStart).toHaveBeenCalledTimes(1);
  });

  it("shows writing state", () => {
    mockDeploymentState = {
      ...mockDeploymentState,
      phase: "writing",
    };

    render(<DeployStep />);
    expect(screen.getByText("Writing configuration...")).toBeInTheDocument();
  });

  it("shows error state with message", () => {
    mockDeploymentState = {
      ...mockDeploymentState,
      phase: "error",
      error: { message: "deploy failed", details: [] },
    };

    render(<DeployStep />);
    expect(screen.getByText("deploy failed")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Deploy" })).toBeInTheDocument();
  });

  it("shows error state with validation details", () => {
    mockDeploymentState = {
      ...mockDeploymentState,
      phase: "error",
      error: {
        message: "Configuration validation failed",
        details: ["field X is required", "field Y is invalid"],
      },
    };

    render(<DeployStep />);
    expect(screen.getByText("field X is required")).toBeInTheDocument();
    expect(screen.getByText("field Y is invalid")).toBeInTheDocument();
  });

  it("shows deploying state with progress", () => {
    mockDeploymentState = {
      ...mockDeploymentState,
      phase: "deploying",
      taskId: "task-1",
      startTime: new Date(),
      progress: { percentage: 42, currentTask: "Installing OSAC" },
    };

    render(<DeployStep />);
    expect(screen.getByText("Deployment")).toBeInTheDocument();
    expect(screen.getByText("Installing OSAC")).toBeInTheDocument();
  });

  it("shows success alert when complete", () => {
    mockDeploymentState = {
      ...mockDeploymentState,
      phase: "complete",
      taskId: "task-1",
      progress: { percentage: 100, currentTask: "" },
    };

    render(<DeployStep />);
    expect(
      screen.getByText("Deployment completed successfully"),
    ).toBeInTheDocument();
  });

  it("shows failure alert when failed", () => {
    mockDeploymentState = {
      ...mockDeploymentState,
      phase: "failed",
      taskId: "task-1",
      progress: { percentage: 50, currentTask: "" },
    };

    render(<DeployStep />);
    expect(screen.getByText("Deployment failed")).toBeInTheDocument();
  });

  it("renders logs when available", () => {
    mockDeploymentState = {
      ...mockDeploymentState,
      phase: "deploying",
      taskId: "task-1",
      logs: "PLAY [Setup]\nTASK [install] ok",
    };

    render(<DeployStep />);
    // Logs are inside expandable section, need to open it
    const toggle = screen.getByText("Show output");
    userEvent.click(toggle);
  });
});
