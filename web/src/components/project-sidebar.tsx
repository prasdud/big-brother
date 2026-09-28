import { Link, useNavigate, useRouterState } from "@tanstack/react-router";
import { Activity } from "lucide-react";
import { useProjects } from "@/lib/project";
import { PROJECT_NAV, isActivePath } from "@/components/nav-items";
import { UserMenu } from "@/components/user-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";

export function ProjectSidebar() {
  const { project, projects, select } = useProjects();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <Link
          to="/"
          className="flex items-center gap-2 px-2 py-1.5 text-foreground hover:opacity-80"
          title="All projects"
        >
          <Activity className="size-4 shrink-0 text-primary" />
          <span className="text-sm font-semibold tracking-tight group-data-[collapsible=icon]:hidden">
            big-brother
          </span>
        </Link>
        <div className="px-1 group-data-[collapsible=icon]:hidden">
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
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {PROJECT_NAV.map((item) => {
                const Icon = item.icon;
                return (
                  <SidebarMenuItem key={item.to}>
                    <SidebarMenuButton asChild isActive={isActivePath(pathname, item)} tooltip={item.label}>
                      <Link to={item.to}>
                        <Icon />
                        <span>{item.label}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <UserMenu />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
