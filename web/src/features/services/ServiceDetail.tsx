import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import { servicesApi } from "../../lib/api";
import { errorMessage, formatDurationMs, formatRelative, formatTime, serviceTarget } from "../../lib/format";
import { useProjects } from "../../lib/project";
import { useSession } from "../../lib/session";
import { ChannelPicker } from "../../components/ChannelPicker";
import { StateBadge } from "../../components/StateBadge";
import { useToast } from "../../components/Toast";
import { Button, Card, CardBody, CardHeader, Spinner } from "../../components/ui";

export function ServiceDetail() {
  const params = useParams({ strict: false }) as { service: string };
  const { project } = useProjects();
  const { canWrite } = useSession();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToast();
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
      toast.push("success", "Service updated");
      void queryClient.invalidateQueries({ queryKey: ["service", slug, params.service] });
      void queryClient.invalidateQueries({ queryKey: ["status", slug, params.service] });
    },
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  const remove = useMutation({
    mutationFn: () => servicesApi.remove(slug, params.service),
    onSuccess: () => {
      toast.push("success", "Service deleted");
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
      void navigate({ to: "/" });
    },
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  if (!project) return <p className="text-sm text-slate-600">Select a project first.</p>;
  if (service.isLoading) return <Spinner />;
  if (service.isError || !service.data) return <p className="text-sm text-red-600">Service not found.</p>;

  const svc = service.data;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-semibold text-slate-900">{svc.name}</h1>
          <StateBadge state={svc.enabled ? status.data?.state : "paused"} />
        </div>
        {canWrite ? (
          <div className="flex gap-2">
            <Button variant="secondary" onClick={() => toggle.mutate()}>
              {svc.enabled ? "Pause" : "Resume"}
            </Button>
            <Button
              variant="secondary"
              onClick={() => void navigate({ to: "/services/$service/edit", params: { service: svc.slug } })}
            >
              Edit
            </Button>
            <Button
              variant="danger"
              onClick={() => {
                if (window.confirm(`Delete ${svc.name}?`)) remove.mutate();
              }}
            >
              Delete
            </Button>
          </div>
        ) : null}
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <Card>
          <CardHeader title="Configuration" />
          <CardBody className="space-y-2 text-sm">
            <Row label="Type" value={svc.type.toUpperCase()} />
            <Row label="Target" value={serviceTarget(svc)} />
            <Row label="Interval" value={`${svc.interval_seconds}s`} />
            <Row label="Timeout" value={`${svc.timeout_seconds}s`} />
            <Row label="Failure threshold" value={String(svc.failure_threshold)} />
          </CardBody>
        </Card>

        <Card>
          <CardHeader title="Current status" />
          <CardBody className="space-y-2 text-sm">
            <Row label="State" value={status.data?.state ?? "unknown"} />
            <Row label="Last check" value={formatRelative(status.data?.last_check_at ?? "")} />
            <Row label="Changed" value={formatTime(status.data?.last_change_at ?? "")} />
            <Row label="Uptime (24h)" value={uptime.data ? `${uptime.data.percent}%` : "—"} />
          </CardBody>
        </Card>
      </div>

      <Card>
        <CardHeader title="Alert channel" />
        <CardBody>
          <ChannelPicker
            project={slug}
            service={svc.slug}
            currentChannelID={svc.channel_id}
            canWrite={canWrite}
          />
        </CardBody>
      </Card>

      <Card>
        <CardHeader title="Recent checks" />
        <CardBody className="p-0">
          {checks.isLoading ? (
            <div className="p-4">
              <Spinner />
            </div>
          ) : (checks.data ?? []).length === 0 ? (
            <p className="p-4 text-sm text-slate-500">No checks yet.</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-200 text-left text-xs uppercase text-slate-500">
                  <th className="px-4 py-2">Time</th>
                  <th className="px-4 py-2">Status</th>
                  <th className="px-4 py-2">Code</th>
                  <th className="px-4 py-2">Latency</th>
                  <th className="px-4 py-2">Error</th>
                </tr>
              </thead>
              <tbody>
                {(checks.data ?? []).map((check) => (
                  <tr key={check.id} className="border-b border-slate-100 last:border-0">
                    <td className="px-4 py-2 text-slate-600">{formatTime(check.checked_at)}</td>
                    <td className="px-4 py-2">
                      <StateBadge state={check.status === "up" ? "up" : "down"} />
                    </td>
                    <td className="px-4 py-2 text-slate-600">{check.status_code || "—"}</td>
                    <td className="px-4 py-2 text-slate-600">{formatDurationMs(check.latency_ms)}</td>
                    <td className="max-w-xs truncate px-4 py-2 text-slate-500">{check.error}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardBody>
      </Card>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4">
      <span className="text-slate-500">{label}</span>
      <span className="truncate text-slate-800">{value}</span>
    </div>
  );
}
