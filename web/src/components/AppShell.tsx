import { useEffect } from "react";
import { Outlet, useNavigate, useRouterState } from "@tanstack/react-router";
import { Plus } from "lucide-react";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { AppSidebar } from "@/components/app-sidebar";
import { useSession } from "@/lib/session";
import { useProjects } from "@/lib/project";

const titles: Record<string, string> = {
  "/": "Projects",
  "/dashboard": "Dashboard",
  "/alerts": "Alerts",
  "/deliveries": "Deliveries",
  "/slack": "Slack",
  "/settings": "Settings",
  "/users": "Users",
  "/services/new": "New service",
};

export function AppShell() {
  const { state, canWrite } = useSession();
  const { project } = useProjects();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  useEffect(() => {
    if (state === "anonymous") void navigate({ to: "/login" });
  }, [state, navigate]);

  if (state === "loading") {
    return (
      <div className="space-y-3 p-8">
        <Skeleton className="h-8 w-56" />
        <Skeleton className="h-40 w-full" />
      </div>
    );
  }
  if (state === "anonymous") return null;

  const page = pathname.startsWith("/services/") ? "Service" : (titles[pathname] ?? "Dashboard");
  const showNew = canWrite && Boolean(project) && pathname !== "/" && pathname !== "/services/new";
  const showProjectCrumb = Boolean(project) && pathname !== "/";

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-14 shrink-0 items-center gap-2 border-b border-border px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-1 h-5" />
          <Breadcrumb>
            <BreadcrumbList>
              {showProjectCrumb ? (
                <>
                  <BreadcrumbItem className="hidden md:block text-muted-foreground">
                    {project?.name}
                  </BreadcrumbItem>
                  <BreadcrumbSeparator className="hidden md:block" />
                </>
              ) : null}
              <BreadcrumbItem>
                <BreadcrumbPage>{page}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          {showNew ? (
            <Button className="ml-auto" size="sm" onClick={() => void navigate({ to: "/services/new" })}>
              <Plus />
              New service
            </Button>
          ) : null}
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 md:p-6">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
