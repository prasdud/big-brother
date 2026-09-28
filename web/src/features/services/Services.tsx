import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Copy, Pause, Pencil, Play, Plus, Search, Trash2, X } from "lucide-react";
import { toast } from "sonner";
import { servicesApi } from "@/lib/api";
import { errorMessage, formatRelative } from "@/lib/format";
import { useProjects } from "@/lib/project";
import { useSession } from "@/lib/session";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { HeartbeatBars, ResponseGraph, Timeline } from "@/features/services/monitor-widgets";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import type { Monitor, State } from "@/lib/types";

const STATE_LABEL: Record<State, string> = {
  up: "Up",
  down: "Down",
  pending: "Pending",
  paused: "Paused",
};

function uptimePillClass(monitor: Monitor): string {
  if (!monitor.enabled) return "bg-muted text-muted-foreground";
  if (monitor.state === "down") return "bg-destructive/15 text-destructive";
  if (monitor.state === "pending") return "bg-muted text-muted-foreground";
  return "bg-primary/15 text-primary";
}

function statePillClass(state: State): string {
  if (state === "up") return "bg-primary/15 text-primary";
  if (state === "down") return "bg-destructive/15 text-destructive";
  return "bg-muted text-muted-foreground";
}

export function Services() {
  const { project } = useProjects();
  const { canWrite } = useSession();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const slug = project?.slug ?? "";

  const [selected, setSelected] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [activeFilter, setActiveFilter] = useState("all");
  const [tagFilter, setTagFilter] = useState("all");
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  const monitors = useQuery({
    queryKey: ["monitors", slug],
    queryFn: () => servicesApi.monitors(slug),
    enabled: Boolean(slug),
    refetchInterval: 10000,
  });

  const list = monitors.data ?? [];
  const tags = useMemo(() => [...new Set(list.flatMap((m) => m.tags))].sort(), [list]);

  const filtered = list.filter((monitor) => {
    if (search && !monitor.name.toLowerCase().includes(search.toLowerCase())) return false;
    if (typeFilter !== "all" && monitor.type !== typeFilter) return false;
    if (statusFilter !== "all" && monitor.state !== statusFilter) return false;
    if (activeFilter === "active" && !monitor.enabled) return false;
    if (activeFilter === "paused" && monitor.enabled) return false;
    if (tagFilter !== "all" && !monitor.tags.includes(tagFilter)) return false;
    return true;
  });

  const current = filtered.find((monitor) => monitor.slug === selected) ?? filtered[0] ?? null;

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ["monitors", slug] });
    void queryClient.invalidateQueries({ queryKey: ["services", slug] });
  };

  const toggle = useMutation({
    mutationFn: (monitor: Monitor) =>
      monitor.enabled ? servicesApi.pause(slug, monitor.slug) : servicesApi.resume(slug, monitor.slug),
    onSuccess: () => {
      toast.success("Monitor updated");
      invalidate();
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  const clone = useMutation({
    mutationFn: (monitor: Monitor) => servicesApi.clone(slug, monitor.slug),
    onSuccess: (service) => {
      toast.success("Monitor cloned (paused)");
      setSelected(service.slug);
      invalidate();
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  const remove = useMutation({
    mutationFn: (monitor: Monitor) => servicesApi.remove(slug, monitor.slug),
    onSuccess: () => {
      toast.success("Monitor deleted");
      setConfirmingDelete(false);
      setSelected(null);
      invalidate();
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  if (!project) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center gap-3 py-10">
          <p className="text-sm text-muted-foreground">No project selected.</p>
          <Button variant="outline" onClick={() => void navigate({ to: "/" })}>
            Go to projects
          </Button>
        </CardContent>
      </Card>
    );
  }

  const clearFilters = () => {
    setSearch("");
    setTypeFilter("all");
    setStatusFilter("all");
    setActiveFilter("all");
    setTagFilter("all");
  };

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        {canWrite ? (
          <Button className="h-11 rounded-2xl px-5" onClick={() => void navigate({ to: "/services/new" })}>
            <Plus />
            Add New Monitor
          </Button>
        ) : (
          <span />
        )}
        <h1 className="text-display text-muted-foreground">{project.name}</h1>
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(360px,34%)_1fr]">
        {/* Monitor list */}
        <Card className="flex max-h-[calc(100svh-11rem)] flex-col overflow-hidden">
          <CardContent className="flex min-h-0 flex-1 flex-col gap-3 p-4">
            <div className="flex gap-2">
              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="http">HTTP</SelectItem>
                  <SelectItem value="tcp">TCP</SelectItem>
                  <SelectItem value="dns">DNS</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  placeholder="Search..."
                  className="pl-9"
                />
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <Button variant="outline" size="icon-sm" onClick={clearFilters} aria-label="Clear filters">
                <X />
              </Button>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="h-7 w-auto gap-1 rounded-full px-3 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Status</SelectItem>
                  <SelectItem value="up">Up</SelectItem>
                  <SelectItem value="down">Down</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="paused">Paused</SelectItem>
                </SelectContent>
              </Select>
              <Select value={activeFilter} onValueChange={setActiveFilter}>
                <SelectTrigger className="h-7 w-auto gap-1 rounded-full px-3 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Active</SelectItem>
                  <SelectItem value="active">Enabled</SelectItem>
                  <SelectItem value="paused">Paused</SelectItem>
                </SelectContent>
              </Select>
              <Select value={tagFilter} onValueChange={setTagFilter}>
                <SelectTrigger className="h-7 w-auto gap-1 rounded-full px-3 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Tags</SelectItem>
                  {tags.map((tag) => (
                    <SelectItem key={tag} value={tag}>
                      {tag}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="-mr-1 min-h-0 flex-1 space-y-1 overflow-y-auto pr-1">
              {monitors.isLoading ? (
                <div className="space-y-2">
                  <Skeleton className="h-14 w-full" />
                  <Skeleton className="h-14 w-full" />
                  <Skeleton className="h-14 w-full" />
                </div>
              ) : filtered.length === 0 ? (
                <p className="py-6 text-center text-sm text-muted-foreground">No monitors.</p>
              ) : (
                filtered.map((monitor) => (
                  <button
                    key={monitor.id}
                    type="button"
                    onClick={() => setSelected(monitor.slug)}
                    className={cn(
                      "flex w-full items-center gap-2 rounded-xl px-2.5 py-2 text-left transition-colors",
                      current?.slug === monitor.slug ? "bg-secondary" : "hover:bg-secondary/60",
                    )}
                  >
                    <span
                      className={cn(
                        "inline-flex min-w-8 shrink-0 items-center justify-center rounded-full px-1 py-0.5 text-[10px] font-medium tabular-nums",
                        uptimePillClass(monitor),
                      )}
                    >
                      {Math.round(monitor.uptime_24h)}%
                    </span>
                    <span className="text-muted-foreground">›</span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm">{monitor.name}</p>
                      {monitor.tags.length > 0 ? (
                        <div className="mt-0.5 flex flex-wrap gap-1">
                          {monitor.tags.map((tag) => (
                            <Badge
                              key={tag}
                              className="border-transparent bg-info/15 text-[10px] text-info"
                            >
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      ) : null}
                    </div>
                    <HeartbeatBars beats={monitor.heartbeats} bars={16} size="sm" />
                  </button>
                ))
              )}
            </div>
          </CardContent>
        </Card>

        {/* Monitor detail */}
        {current ? (
          <div className="space-y-4">
            <div className="flex flex-wrap gap-2">
              {canWrite ? (
                <>
                  <Button
                    variant="secondary"
                    className="rounded-full"
                    onClick={() => toggle.mutate(current)}
                  >
                    {current.enabled ? <Pause /> : <Play />}
                    {current.enabled ? "Pause" : "Resume"}
                  </Button>
                  <Button
                    variant="secondary"
                    className="rounded-full"
                    onClick={() =>
                      void navigate({
                        to: "/services/$service/edit",
                        params: { service: current.slug },
                      })
                    }
                  >
                    <Pencil />
                    Edit
                  </Button>
                  <Button
                    variant="secondary"
                    className="rounded-full"
                    onClick={() => clone.mutate(current)}
                  >
                    <Copy />
                    Clone
                  </Button>
                  <Button
                    variant="destructive"
                    className="rounded-full"
                    onClick={() => setConfirmingDelete(true)}
                  >
                    <Trash2 />
                    Delete
                  </Button>
                </>
              ) : null}
            </div>

            <Card>
              <CardContent className="space-y-4 p-6">
                <div className="flex items-start gap-4">
                  <div className="min-w-0 flex-1">
                    <Timeline beats={current.heartbeats} />
                    <div className="mt-2 flex justify-between text-xs text-muted-foreground">
                      <span>
                        {current.heartbeats.length
                          ? formatRelative(current.heartbeats[0].checked_at)
                          : "—"}
                      </span>
                      <span>now</span>
                    </div>
                  </div>
                  <div
                    className={cn(
                      "shrink-0 rounded-2xl px-6 py-3 text-lg font-semibold",
                      statePillClass(current.state),
                    )}
                  >
                    {STATE_LABEL[current.state]}
                  </div>
                </div>
                <p className="text-sm text-muted-foreground">
                  Check every <span className="text-foreground">{current.interval_seconds}</span> seconds
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="grid grid-cols-2 divide-x divide-border py-8">
                <div className="px-6 text-center">
                  <p className="text-sm text-muted-foreground">Uptime</p>
                  <p className="text-xs text-muted-foreground">(24-hour)</p>
                  <p className="mt-3 text-display tabular-nums text-primary">{current.uptime_24h}%</p>
                </div>
                <div className="px-6 text-center">
                  <p className="text-sm text-muted-foreground">Uptime</p>
                  <p className="text-xs text-muted-foreground">(30-day)</p>
                  <p className="mt-3 text-display tabular-nums text-primary">{current.uptime_30d}%</p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm text-muted-foreground">Response Time (ms)</CardTitle>
                <div className="ml-auto">
                  <Select value="recent" onValueChange={() => {}}>
                    <SelectTrigger className="h-8 w-28">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="recent">Recent</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </CardHeader>
              <CardContent>
                <ResponseGraph beats={current.heartbeats} />
              </CardContent>
            </Card>
          </div>
        ) : (
          <Card>
            <CardContent className="py-16 text-center text-sm text-muted-foreground">
              {monitors.isLoading ? "Loading…" : "Add a monitor to get started."}
            </CardContent>
          </Card>
        )}
      </div>

      <ConfirmDialog
        open={confirmingDelete}
        onOpenChange={setConfirmingDelete}
        title={`Delete ${current?.name ?? "monitor"}?`}
        description="This removes the monitor and its check history."
        confirmLabel="Delete"
        pending={remove.isPending}
        onConfirm={() => {
          if (current) remove.mutate(current);
        }}
      />
    </div>
  );
}
