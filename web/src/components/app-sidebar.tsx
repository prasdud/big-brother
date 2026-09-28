import { Link, useNavigate, useRouterState } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Activity, Blocks, FolderKanban, LogOut, Users } from "lucide-react";
import { useSession } from "@/lib/session";
import { useProjects } from "@/lib/project";
import { slackApi } from "@/lib/api";
import { cn } from "@/lib/utils";
import { PROJECT_NAV } from "@/components/project-nav";
import { Badge } from "@/components/ui/badge";
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
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

const NAV = [
  { to: "/", label: "Projects", icon: FolderKanban, exact: true, adminOnly: false },
  { to: "/integrations", label: "Integrations", icon: Blocks, exact: false, adminOnly: false },
  { to: "/users", label: "Users", icon: Users, exact: false, adminOnly: true },
] as const;

export function AppSidebar() {
  const { user, role, isAdmin, signOut } = useSession();
  const { project, projects, select } = useProjects();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const slack = useQuery({ queryKey: ["slack"], queryFn: slackApi.status });

  const initials = (user?.email ?? "dev").slice(0, 2).toUpperCase();

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="flex items-center gap-2 px-2 py-1.5">
          <Activity className="size-4 shrink-0 text-primary" />
          <span className="text-sm font-semibold tracking-tight group-data-[collapsible=icon]:hidden">
            big-brother
          </span>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {NAV.filter((item) => !item.adminOnly || isAdmin).map((item) => {
                const Icon = item.icon;
                const active = item.exact ? pathname === item.to : pathname.startsWith(item.to);
                return (
                  <SidebarMenuItem key={item.to}>
                    <SidebarMenuButton asChild isActive={active} tooltip={item.label}>
                      <Link to={item.to}>
                        <Icon />
                        <span>{item.label}</span>
                        {item.to === "/integrations" ? (
                          <Badge
                            variant="outline"
                            className={cn(
                              "ml-auto group-data-[collapsible=icon]:hidden",
                              slack.data?.connected
                                ? "border-primary/40 bg-primary/10 text-primary"
                                : "text-muted-foreground",
                            )}
                          >
                            Slack
                          </Badge>
                        ) : null}
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        {project ? (
          <SidebarGroup className="md:hidden">
            <SidebarGroupLabel>{project.name}</SidebarGroupLabel>
            <SidebarGroupContent>
              <div className="px-1 pb-2">
                <Select
                  value={project.slug}
                  onValueChange={(value) => {
                    select(value);
                    void navigate({ to: "/services" });
                  }}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
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
              <SidebarMenu>
                {PROJECT_NAV.map((item) => {
                  const Icon = item.icon;
                  const active =
                    item.to === "/services"
                      ? pathname === "/services" || pathname.startsWith("/services/")
                      : pathname === item.to;
                  return (
                    <SidebarMenuItem key={item.to}>
                      <SidebarMenuButton asChild isActive={active}>
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
        ) : null}
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton size="lg">
                  <Avatar className="size-7 rounded-md">
                    <AvatarFallback className="rounded-md text-xs">{initials}</AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-xs leading-tight group-data-[collapsible=icon]:hidden">
                    <span className="truncate font-medium">{user?.email ?? "local dev"}</span>
                    <span className="truncate text-muted-foreground">{role}</span>
                  </div>
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top" align="start" className="w-56">
                <DropdownMenuLabel>{user?.email ?? "local dev"}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => {
                    void signOut().then(() => navigate({ to: "/login" }));
                  }}
                >
                  <LogOut />
                  Sign out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
