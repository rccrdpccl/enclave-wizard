import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { DeployStep } from "./DeployStep.tsx";

const mockWriteConfig = vi.fn();
const mockStartDeploy = vi.fn();
const mockGetTask = vi.fn();
const mockGetTaskLogs = vi.fn();

vi.mock("../../api/useEnclaveApi.ts", () => ({
  useEnclaveApi: () => ({ writeConfig: mockWriteConfig }),
}));

vi.mock("../../api/useTasksApi.ts", () => ({
  useTasksApi: () => ({
    startDeploy: mockStartDeploy,
    getTask: mockGetTask,
    getTaskLogs: mockGetTaskLogs,
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
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 404 }),
    );
    Element.prototype.scrollIntoView = vi.fn();
  });

  it("shows deploy button in idle state", () => {
    render(<DeployStep />);
    expect(screen.getByRole("button", { name: "Deploy" })).toBeInTheDocument();
  });

  it("calls writeConfig then startDeploy on click", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    mockStartDeploy.mockResolvedValue({ id: "run-1", status: "running" });

    render(<DeployStep />);
    await userEvent.click(screen.getByRole("button", { name: "Deploy" }));

    expect(mockWriteConfig).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(mockStartDeploy).toHaveBeenCalledTimes(1));
  });

  it("shows error page when startDeploy fails", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    mockStartDeploy.mockRejectedValue(new Error("deploy failed"));

    render(<DeployStep />);
    await userEvent.click(screen.getByRole("button", { name: "Deploy" }));

    await waitFor(() =>
      expect(screen.getByText("deploy failed")).toBeInTheDocument(),
    );
    expect(
      screen.getByRole("button", { name: "Deploy" }),
    ).toBeInTheDocument();
  });

  it("shows error page when writeConfig fails", async () => {
    mockWriteConfig.mockRejectedValue(new Error("write failed"));

    render(<DeployStep />);
    await userEvent.click(screen.getByRole("button", { name: "Deploy" }));

    await waitFor(() =>
      expect(screen.getByText("write failed")).toBeInTheDocument(),
    );
    expect(
      screen.getByRole("button", { name: "Deploy" }),
    ).toBeInTheDocument();
  });

  it("enters deploying state and polls task logs", async () => {
    mockWriteConfig.mockResolvedValue(undefined);
    mockStartDeploy.mockResolvedValue({ id: "run-1", status: "running" });
    mockGetTask.mockResolvedValue({ id: "run-1", status: "running" });
    mockGetTaskLogs.mockResolvedValue("PLAY [Prepare] ***");

    render(<DeployStep />);
    await userEvent.click(screen.getByRole("button", { name: "Deploy" }));

    await waitFor(() =>
      expect(screen.getByText("Deployment")).toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(mockGetTaskLogs).toHaveBeenCalledWith("run-1"),
    );
  });
});
