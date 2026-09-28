import { useQuery } from "@tanstack/react-query";
import { alertsApi } from "@/lib/api";
import { formatTime } from "@/lib/format";
import { useProjects } from "@/lib/project";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function Deliveries() {
  const { project } = useProjects();
  const deliveries = useQuery({
    queryKey: ["deliveries", project?.slug],
    queryFn: () => alertsApi.deliveries(project!.slug),
    enabled: Boolean(project),
    refetchInterval: 15000,
  });

  if (!project) return <p className="text-sm text-muted-foreground">Select a project first.</p>;

  const rows = deliveries.data ?? [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Alert delivery failures</CardTitle>
        <CardDescription>Alerts that could not be delivered.</CardDescription>
      </CardHeader>
      <CardContent className="px-0">
        {deliveries.isLoading ? (
          <div className="space-y-2 px-6">
            <Skeleton className="h-10 w-full" />
          </div>
        ) : rows.length === 0 ? (
          <p className="px-6 text-sm text-muted-foreground">No delivery failures. Nice.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Time</TableHead>
                <TableHead>Trigger</TableHead>
                <TableHead>Reason</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((delivery) => (
                <TableRow key={delivery.id}>
                  <TableCell className="text-muted-foreground">{formatTime(delivery.created_at)}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-destructive">
                      {delivery.trigger || "—"}
                    </Badge>
                  </TableCell>
                  <TableCell>{delivery.reason}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
