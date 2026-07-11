// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createElement } from "react";
import { OwnershipActions } from "./ownership-actions";

afterEach(cleanup);

describe("OwnershipActions", () => {
  it("never renders edit/delete for a post the server says is not owned", () => {
    render(
      createElement(OwnershipActions, {
        viewerCanEdit: false,
        onEdit: vi.fn(),
        onDelete: vi.fn(),
      }),
    );
    expect(screen.queryByRole("button", { name: "Edit" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Delete" })).toBeNull();
  });

  it("renders accessible actions and delegates for an owned post", () => {
    const onEdit = vi.fn();
    render(
      createElement(OwnershipActions, {
        viewerCanEdit: true,
        onEdit,
        onDelete: vi.fn(),
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Edit" }));
    expect(onEdit).toHaveBeenCalledOnce();
  });
});
