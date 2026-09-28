import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import { toast } from "sonner";
import { servicesApi } from "@/lib/api";
import { errorMessage, formatDurationMs, formatRelative, formatTime, serviceTarget } from "@/lib/format";
import { useProjects } from "@/lib/project";
import { useSession } from "@/lib/session";
import { ChannelPicker } from "@/components/ChannelPicker";
import { StatusBadge } from "@/components/status-badge";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-border/60 py-2 last:border-0">
      <span className="text-label uppercase text-muted-foreground">{label}</span>
      <span className="truncate text-sm">{value}</span>
    </div>
  );
}

export function ServiceDetail() {
  const params = useParams({ strict: false }) as { service: string };
  const { project } = useProjects();
  const { canWrite } = useSession();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [confirming, setConfirming] = useState(false);
  const slug = project?.slug ?? "";

  const service = useQuery({
    queryKey: ["service", slug, params.service],
    queryFn: () => servicesApi.get(slug, params.service),
    enabled: Boolean(slug),
  });
  const status = useQuery({
    queryKey: ["status", slug, params.service],
    queryFn: () => servicesApi.status(slug, params.service),
    enabled: Boolean(slug),
    refetchInterval: 5000,
  });
  const checks = useQuery({
    queryKey: ["checks", slug, params.service],
    queryFn: () => servicesApi.checks(slug, params.service, 50),
    enabled: Boolean(slug),
    refetchInterval: 10000,
  });
  const uptime = useQuery({
    queryKey: ["uptime", slug, params.service],
    queryFn: () => servicesApi.uptime(slug, params.service, "24h"),
    enabled: Boolean(slug),
    refetchInterval: 30000,
  });

  const toggle = useMutation({
    mutationFn: () =>
      service.data?.enabled
        ? servicesApi.pause(slug, params.service)
        : servicesApi.resume(slug, params.service),
    onSuccess: () => {
      toast.success("Service updated");
      void queryClient.invalidateQueries({ queryKey: ["service", slug, params.service] });
      void queryClient.invalidateQueries({ queryKey: ["status", slug, params.service] });
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  const remove = useMutation({
    mutationFn: () => servicesApi.remove(slug, params.service),
    onSuccess: () => {
      toast.success("Service deleted");
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
      void navigate({ to: "/dashboard" });
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  if (!project) return <p className="text-sm text-muted-foreground">Select a project first.</p>;
  if (service.isLoading) return <Skeleton className="h-96 w-full" />;
  if (service.isError || !service.data) return <p className="text-sm text-destructive">Service not found.</p>;

  const svc = service.data;
  const percent = uptime.data?.percent ?? 0;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <h1 className="text-h1">{svc.name}</h1>
          <StatusBadge state={svc.enabled ? status.data?.state : "paused"} />
          <Badge variant="outline" className="uppercase text-muted-foreground">
            {svc.type}
          </Badge>
        </div>
        {canWrite ? (
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => toggle.mutate()}>
              {svc.enabled ? "Pause" : "Resume"}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void navigate({ to: "/services/$service/edit", params: { service: svc.slug } })}
            >
              Edit
            </Button>
            <Button variant="destructive" size="sm" onClick={() => setConfirming(true)}>
              Delete
            </Button>
          </div>
        ) : null}
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="checks">Checks</TabsTrigger>
          <TabsTrigger value="uptime">Uptime</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-4 space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Configuration</CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <Row label="Target" value={serviceTarget(svc)} />
                <Row label="Interval" value={`${svc.interval_seconds}s`} />
                <Row label="Timeout" value={`${svc.timeout_seconds}s`} />
                <Row label="Failure threshold" value={String(svc.failure_threshold)} />
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Current status</CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <Row label="State" value={status.data?.state ?? "unknown"} />
                <Row label="Last check" value={formatRelative(status.data?.last_check_at ?? "")} />
                <Row label="Changed" value={formatTime(status.data?.last_change_at ?? "")} />
                <Row label="Consecutive failures" value={String(status.data?.consecutive_failures ?? 0)} />
              </CardContent>
            </Card>
          </div>
          <Card>
            <CardHeader>
              <CardTitle>Alert channel</CardTitle>
            </CardHeader>
            <CardContent>
              <ChannelPicker
                project={slug}
                service={svc.slug}
                currentChannelID={svc.channel_id}
                canWrite={canWrite}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="checks" className="mt-4">
          <Card>
            <CardContent className="px-0">
              {checks.isLoading ? (
                <div className="space-y-2 px-6">
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-10 w-full" />
                </div>
              ) : (checks.data ?? []).length === 0 ? (
                <p className="px-6 text-sm text-muted-foreground">No checks yet.</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Time</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Code</TableHead>
                      <TableHead>Latency</TableHead>
                      <TableHead>Error</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(checks.data ?? []).map((check) => (
                      <TableRow key={check.id}>
                        <TableCell className="text-muted-foreground">{formatTime(check.checked_at)}</TableCell>
                        <TableCell>
                          <StatusBadge state={check.status === "up" ? "up" : "down"} />
                        </TableCell>
                        <TableCell className="tabular-nums text-muted-foreground">
                          {check.status_code || "—"}
                        </TableCell>
                        <TableCell className="tabular-nums text-muted-foreground">
                          {formatDurationMs(check.latency_ms)}
                        </TableCell>
                        <TableCell className="max-w-xs truncate text-muted-foreground">
                          {check.error}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="uptime" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Uptime · 24h</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-display tabular-nums">{percent}%</p>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div className="h-full bg-primary" style={{ width: `${Math.min(percent, 100)}%` }} />
              </div>
              <p className="text-label text-muted-foreground">
                {uptime.data?.up_checks ?? 0} up of {uptime.data?.total_checks ?? 0} checks
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        open={confirming}
        onOpenChange={setConfirming}
        title={`Delete ${svc.name}?`}
        description="This removes the service and its check history."
        confirmLabel="Delete"
        pending={remove.isPending}
        onConfirm={() => remove.mutate()}
      />
    </div>
  );
}
