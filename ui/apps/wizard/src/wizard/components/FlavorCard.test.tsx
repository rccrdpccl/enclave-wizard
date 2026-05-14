import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { FlavorCard } from "./FlavorCard.tsx";

describe("FlavorCard", () => {
  it("renders the flavor title and description", () => {
    render(
      <FlavorCard
        title="Cluster as a Service"
        description="On-demand container clusters."
        isSelected={false}
        onSelect={vi.fn()}
      />,
    );
    expect(screen.getByText("Cluster as a Service")).toBeInTheDocument();
    expect(
      screen.getByText("On-demand container clusters."),
    ).toBeInTheDocument();
  });

  it("shows selected state", () => {
    render(
      <FlavorCard
        title="Cluster as a Service"
        description="Test"
        isSelected={true}
        onSelect={vi.fn()}
      />,
    );
    const card = screen.getByRole("button");
    expect(card).toHaveAttribute("aria-pressed", "true");
  });

  it("calls onSelect when clicked", async () => {
    const onSelect = vi.fn();
    render(
      <FlavorCard
        title="Cluster as a Service"
        description="Test"
        isSelected={false}
        onSelect={onSelect}
      />,
    );
    await userEvent.click(screen.getByRole("button"));
    expect(onSelect).toHaveBeenCalledTimes(1);
  });
});
