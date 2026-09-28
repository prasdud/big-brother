import { useQuery } from "@tanstack/react-query";
import { alertsApi } from "../../lib/api";
import { formatTime } from "../../lib/format";
import { useProjects } from "../../lib/project";
import { Card, CardBody, CardHeader, Spinner } from "../../components/ui";

export function Deliveries() {
  const { project } = useProjects();
  const deliveries = useQuery({
    queryKey: ["deliveries", project?.slug],
    queryFn: () => alertsApi.deliveries(project!.slug),
    enabled: Boolean(project),
    refetchInterval: 15000,
  });

  if (!project) return <p className="text-sm text-slate-600">Select a project first.</p>;

  const rows = deliveries.data ?? [];

  return (
    <Card>
      <CardHeader title="Alert delivery failures" />
      <CardBody className="p-0">
        {deliveries.isLoading ? (
          <div className="p-4">
            <Spinner />
          </div>
        ) : rows.length === 0 ? (
          <p className="p-4 text-sm text-slate-500">No delivery failures. Nice.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-200 text-left text-xs uppercase text-slate-500">
                <th className="px-4 py-2">Time</th>
                <th className="px-4 py-2">Trigger</th>
                <th className="px-4 py-2">Reason</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((delivery) => (
                <tr key={delivery.id} className="border-b border-slate-100 last:border-0">
                  <td className="px-4 py-2 text-slate-600">{formatTime(delivery.created_at)}</td>
                  <td className="px-4 py-2 text-slate-600">{delivery.trigger || "—"}</td>
                  <td className="px-4 py-2 text-slate-800">{delivery.reason}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </CardBody>
    </Card>
  );
}
