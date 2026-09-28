import { Link, useNavigate, useRouterState } from "@tanstack/react-router";
import { Bell, Inbox, Server, Settings } from "lucide-react";
import { useProjects } from "@/lib/project";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

export const PROJECT_NAV = [
  { to: "/services", label: "Services", icon: Server },
  { to: "/alerts", label: "Alerts", icon: Bell },
  { to: "/deliveries", label: "Deliveries", icon: Inbox },
  { to: "/settings", label: "Settings", icon: Settings },
] as const;

function isActive(pathname: string, to: string): boolean {
  if (to === "/services") return pathname === "/services" || pathname.startsWith("/services/");
  return pathname === to;
}

export function ProjectNav() {
  const { project, projects, select } = useProjects();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });

  return (
    <div className="hidden w-56 shrink-0 flex-col gap-2 border-r border-sidebar-border bg-sidebar p-2 md:flex">
      <Select
        value={project?.slug ?? ""}
        onValueChange={(value) => {
          select(value);
          void navigate({ to: "/services" });
        }}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder="Select project" />
        </SelectTrigger>
        <SelectContent>
          {projects.map((item) => (
            <SelectItem key={item.slug} value={item.slug}>
              {item.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <SidebarMenu>
        {PROJECT_NAV.map((item) => {
          const Icon = item.icon;
          return (
            <SidebarMenuItem key={item.to}>
              <SidebarMenuButton asChild isActive={isActive(pathname, item.to)}>
                <Link to={item.to}>
                  <Icon />
                  <span>{item.label}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          );
        })}
      </SidebarMenu>
    </div>
  );
}
