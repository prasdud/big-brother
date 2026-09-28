import { useState } from "react";
import { useMutation, useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { MoreHorizontal } from "lucide-react";
import { toast } from "sonner";
import { servicesApi } from "@/lib/api";
import { errorMessage, serviceTarget } from "@/lib/format";
import { useProjects } from "@/lib/project";
import { useSession } from "@/lib/session";
import { StatusBadge } from "@/components/status-badge";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Service } from "@/lib/types";

function Stat({ label, value }: { label: string; value: number | string }) {
  return (
    <Card className="py-4">
      <CardContent className="space-y-1">
        <p className="text-label uppercase text-muted-foreground">{label}</p>
        <p className="text-h1 tabular-nums">{value}</p>
      </CardContent>
    </Card>
  );
}

export function Dashboard() {
  const { project, loading } = useProjects();
  const { canWrite } = useSession();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [pendingDelete, setPendingDelete] = useState<Service | null>(null);

  const slug = project?.slug ?? "";
  const services = useQuery({
    queryKey: ["services", slug],
    queryFn: () => servicesApi.list(slug),
    enabled: Boolean(slug),
  });

  const statuses = useQueries({
    queries: (services.data ?? []).map((service) => ({
      queryKey: ["status", slug, service.slug],
      queryFn: () => servicesApi.status(slug, service.slug),
      refetchInterval: 5000,
    })),
  });

  const toggle = useMutation({
    mutationFn: ({ name, enabled }: { name: string; enabled: boolean }) =>
      enabled ? servicesApi.pause(slug, name) : servicesApi.resume(slug, name),
    onSuccess: () => {
      toast.success("Service updated");
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  const remove = useMutation({
    mutationFn: (name: string) => servicesApi.remove(slug, name),
    onSuccess: () => {
      toast.success("Service deleted");
      setPendingDelete(null);
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  if (loading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-56 w-full" />
      </div>
    );
  }
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

  const rows = services.data ?? [];
  const counts = { up: 0, down: 0, pending: 0, paused: 0 };
  rows.forEach((service, index) => {
    if (!service.enabled) {
      counts.paused += 1;
      return;
    }
    const state = statuses[index]?.data?.state;
    if (state && state in counts) counts[state as keyof typeof counts] += 1;
  });

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-h1">{project.name}</h1>
        <p className="text-sm text-muted-foreground">Current status across your services.</p>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <Stat label="Up" value={counts.up} />
        <Stat label="Down" value={counts.down} />
        <Stat label="Pending" value={counts.pending} />
        <Stat label="Paused" value={counts.paused} />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Services</CardTitle>
          <CardDescription>{rows.length} total</CardDescription>
        </CardHeader>
        <CardContent className="px-0">
          {services.isLoading ? (
            <div className="space-y-2 px-6">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : rows.length === 0 ? (
            <p className="px-6 pb-2 text-sm text-muted-foreground">No services yet.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Service</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((service, index) => (
                  <TableRow key={service.id}>
                    <TableCell>
                      <Link
                        className="font-medium hover:underline"
                        to="/services/$service"
                        params={{ service: service.slug }}
                      >
                        {service.name}
                      </Link>
                      <Badge variant="outline" className="ml-2 uppercase text-muted-foreground">
                        {service.type}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-xs truncate text-muted-foreground">
                      {serviceTarget(service)}
                    </TableCell>
                    <TableCell>
                      <StatusBadge state={service.enabled ? statuses[index]?.data?.state : "paused"} />
                    </TableCell>
                    <TableCell className="text-right">
                      {canWrite ? (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon-sm" aria-label="Actions">
                              <MoreHorizontal />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => toggle.mutate({ name: service.slug, enabled: service.enabled })}
                            >
                              {service.enabled ? "Pause" : "Resume"}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() =>
                                void navigate({
                                  to: "/services/$service/edit",
                                  params: { service: service.slug },
                                })
                              }
                            >
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              variant="destructive"
                              onClick={() => setPendingDelete(service)}
                            >
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      ) : (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open) setPendingDelete(null);
        }}
        title={`Delete ${pendingDelete?.name ?? "service"}?`}
        description="This removes the service and its check history."
        confirmLabel="Delete"
        pending={remove.isPending}
        onConfirm={() => {
          if (pendingDelete) remove.mutate(pendingDelete.slug);
        }}
      />
    </div>
  );
}
