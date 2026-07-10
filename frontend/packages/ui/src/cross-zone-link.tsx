import type { AnchorHTMLAttributes, ReactNode } from "react";

export interface CrossZoneLinkProps extends AnchorHTMLAttributes<HTMLAnchorElement> {
  children: ReactNode;
}

export function CrossZoneLink({ children, ...props }: CrossZoneLinkProps) {
  return <a {...props}>{children}</a>;
}
