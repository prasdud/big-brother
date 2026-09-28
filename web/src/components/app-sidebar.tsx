import { Link, useNavigate, useRouterState } from "@tanstack/react-router";
import {
  Activity,
  Bell,
  FolderKanban,
  Inbox,
  LayoutDashboard,
  LogOut,
  MessageSquare,
  Settings,
  Users,
} from "lucide-react";
import { useSession } from "@/lib/session";
import { useProjects } from "@/lib/project";
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
  { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard, exact: true, adminOnly: false },
  { to: "/alerts", label: "Alerts", icon: Bell, exact: false, adminOnly: false },
  { to: "/deliveries", label: "Deliveries", icon: Inbox, exact: false, adminOnly: false },
  { to: "/slack", label: "Slack", icon: MessageSquare, exact: false, adminOnly: false },
  { to: "/settings", label: "Settings", icon: Settings, exact: false, adminOnly: false },
  { to: "/users", label: "Users", icon: Users, exact: false, adminOnly: true },
] as const;

export function AppSidebar() {
  const { user, role, isAdmin, signOut } = useSession();
  const { projects, project, select } = useProjects();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });

  const initials = (user?.email ?? "dev").slice(0, 2).toUpperCase();

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="flex items-center gap-2 px-2 py-1.5 group-data-[collapsible=icon]:hidden">
          <Activity className="size-4 text-primary" />
          <span className="text-sm font-semibold tracking-tight">big-brother</span>
        </div>
        {projects.length > 0 ? (
          <div className="px-1 group-data-[collapsible=icon]:hidden">
            <Select
              value={project?.slug ?? ""}
              onValueChange={(value) => {
                select(value);
                void navigate({ to: "/dashboard" });
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
        ) : null}
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
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton size="lg">
                  <Avatar className="size-7 rounded-md">
                    <AvatarFallback className="rounded-md text-xs">{initials}</AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-xs leading-tight">
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
