import type { LucideIcon } from "lucide-react";
import { Bell, Blocks, FolderKanban, Inbox, Server, Settings, Users } from "lucide-react";

export interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  exact?: boolean;
  adminOnly?: boolean;
}

export const APP_NAV: NavItem[] = [
  { to: "/", label: "Projects", icon: FolderKanban, exact: true },
  { to: "/integrations", label: "Integrations", icon: Blocks },
  { to: "/users", label: "Users", icon: Users, adminOnly: true },
];

export const PROJECT_NAV: NavItem[] = [
  { to: "/services", label: "Services", icon: Server },
  { to: "/alerts", label: "Alerts", icon: Bell },
  { to: "/deliveries", label: "Deliveries", icon: Inbox },
  { to: "/settings", label: "Settings", icon: Settings },
];

const PROJECT_ROUTES = ["/services", "/alerts", "/deliveries", "/settings"];

/** True when the current location belongs to a project context. */
export function isProjectRoute(pathname: string): boolean {
  return PROJECT_ROUTES.some((route) => pathname === route || pathname.startsWith(`${route}/`));
}

/** Active state for a nav item. */
export function isActivePath(pathname: string, item: NavItem): boolean {
  if (item.exact) return pathname === item.to;
  if (item.to === "/services") return pathname === "/services" || pathname.startsWith("/services/");
  return pathname === item.to;
}
