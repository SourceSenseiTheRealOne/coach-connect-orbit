import { createElement } from "react";

export function OwnershipActions({
  viewerCanEdit,
  onEdit,
  onDelete,
}: {
  viewerCanEdit: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  if (!viewerCanEdit) return null;
  return createElement(
    "span",
    { className: "flex gap-1" },
    createElement(
      "button",
      { className: "min-h-11 px-3 underline", type: "button", onClick: onEdit },
      "Edit",
    ),
    createElement(
      "button",
      {
        className: "min-h-11 px-3 underline",
        type: "button",
        onClick: onDelete,
      },
      "Delete",
    ),
  );
}
