import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { WizardFooter } from "./WizardFooter.tsx";

describe("WizardFooter", () => {
  it("renders Back and Next buttons", () => {
    render(
      <WizardFooter
        onBack={vi.fn()}
        onNext={vi.fn()}
        isFirst={false}
        isLast={false}
        isContinue={false}
      />,
    );
    expect(screen.getByRole("button", { name: "Back" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Next" })).toBeInTheDocument();
  });

  it("disables Back on first step", () => {
    render(
      <WizardFooter
        onBack={vi.fn()}
        onNext={vi.fn()}
        isFirst={true}
        isLast={false}
        isContinue={false}
      />,
    );
    expect(screen.getByRole("button", { name: "Back" })).toBeDisabled();
  });

  it("disables Next on last step", () => {
    render(
      <WizardFooter
        onBack={vi.fn()}
        onNext={vi.fn()}
        isFirst={false}
        isLast={true}
        isContinue={false}
      />,
    );
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
  });

  it("shows Continue label when isContinue is true", () => {
    render(
      <WizardFooter
        onBack={vi.fn()}
        onNext={vi.fn()}
        isFirst={false}
        isLast={false}
        isContinue={true}
      />,
    );
    expect(
      screen.getByRole("button", { name: "Continue" }),
    ).toBeInTheDocument();
  });

  it("calls onBack when Back is clicked", async () => {
    const onBack = vi.fn();
    render(
      <WizardFooter
        onBack={onBack}
        onNext={vi.fn()}
        isFirst={false}
        isLast={false}
        isContinue={false}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Back" }));
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  it("calls onNext when Next is clicked", async () => {
    const onNext = vi.fn();
    render(
      <WizardFooter
        onBack={vi.fn()}
        onNext={onNext}
        isFirst={false}
        isLast={false}
        isContinue={false}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(onNext).toHaveBeenCalledTimes(1);
  });
});
