import { useEffect } from "react";
import { Link, Outlet, useNavigate } from "@tanstack/react-router";
import { useSession } from "../lib/session";
import { useProjects } from "../lib/project";
import { Button, Select, Spinner } from "./ui";

function NavLink({ to, label }: { to: string; label: string }) {
  return (
    <Link
      to={to}
      className="rounded-md px-2 py-1 text-sm text-slate-600 hover:bg-slate-100"
      activeProps={{ className: "rounded-md px-2 py-1 text-sm font-medium bg-slate-200 text-slate-900" }}
      activeOptions={{ exact: to === "/" }}
    >
      {label}
    </Link>
  );
}

export function AppShell() {
  const { state, user, role, isAdmin, signOut } = useSession();
  const { projects, project, select } = useProjects();
  const navigate = useNavigate();

  useEffect(() => {
    if (state === "anonymous") void navigate({ to: "/login" });
  }, [state, navigate]);

  if (state === "loading") {
    return (
      <div className="p-8">
        <Spinner />
      </div>
    );
  }
  if (state === "anonymous") return null;

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-5xl flex-wrap items-center gap-3 px-4 py-3">
          <span className="text-sm font-semibold text-slate-900">big-brother</span>

          {projects.length > 0 ? (
            <Select
              className="w-48"
              value={project?.slug ?? ""}
              onChange={(e) => {
                select(e.target.value);
                void navigate({ to: "/" });
              }}
            >
              {projects.map((p) => (
                <option key={p.slug} value={p.slug}>
                  {p.name}
                </option>
              ))}
            </Select>
          ) : null}

          <nav className="flex items-center gap-1">
            <NavLink to="/" label="Dashboard" />
            <NavLink to="/alerts" label="Alerts" />
            <NavLink to="/settings" label="Settings" />
            <NavLink to="/deliveries" label="Deliveries" />
            <NavLink to="/slack" label="Slack" />
            {isAdmin ? <NavLink to="/users" label="Users" /> : null}
          </nav>

          <div className="ml-auto flex items-center gap-3">
            <span className="text-xs text-slate-500">
              {user?.email ?? "local dev"} · {role}
            </span>
            {state === "disabled" ? null : (
              <Button
                variant="secondary"
                onClick={() => {
                  void signOut().then(() => navigate({ to: "/login" }));
                }}
              >
                Sign out
              </Button>
            )}
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
