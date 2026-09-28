import { useState } from "react";
import { useMutation, useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { projectsApi, servicesApi } from "../../lib/api";
import { errorMessage, serviceTarget } from "../../lib/format";
import { useProjects } from "../../lib/project";
import { useSession } from "../../lib/session";
import { StateBadge } from "../../components/StateBadge";
import { useToast } from "../../components/Toast";
import { Button, Card, CardBody, CardHeader, ErrorText, Field, Input, Spinner } from "../../components/ui";

function CreateProject() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const { select } = useProjects();
  const [name, setName] = useState("");
  const [error, setError] = useState("");

  const create = useMutation({
    mutationFn: () => projectsApi.create(name),
    onSuccess: (project) => {
      toast.push("success", "Project created");
      select(project.slug);
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
    onError: (err) => setError(errorMessage(err)),
  });

  return (
    <Card className="mx-auto max-w-md">
      <CardHeader title="Create your first project" />
      <CardBody>
        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            setError("");
            create.mutate();
          }}
        >
          <Field label="Project name">
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Payments" />
          </Field>
          <ErrorText>{error}</ErrorText>
          <Button type="submit" disabled={create.isPending || !name}>
            Create project
          </Button>
        </form>
      </CardBody>
    </Card>
  );
}

export function Dashboard() {
  const { project, loading } = useProjects();
  const { canWrite, isAdmin } = useSession();
  const queryClient = useQueryClient();
  const toast = useToast();
  const navigate = useNavigate();

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
      toast.push("success", "Service updated");
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
    },
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  const remove = useMutation({
    mutationFn: (name: string) => servicesApi.remove(slug, name),
    onSuccess: () => {
      toast.push("success", "Service deleted");
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
    },
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  if (loading) return <Spinner />;
  if (!project) {
    return isAdmin ? <CreateProject /> : <p className="text-sm text-slate-600">No projects yet. Ask an admin to create one.</p>;
  }

  const rows = services.data ?? [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold text-slate-900">{project.name}</h1>
        {canWrite ? (
          <Button onClick={() => void navigate({ to: "/services/new" })}>New service</Button>
        ) : null}
      </div>

      <Card>
        <CardHeader title="Services" />
        <CardBody className="p-0">
          {services.isLoading ? (
            <div className="p-4">
              <Spinner />
            </div>
          ) : rows.length === 0 ? (
            <p className="p-4 text-sm text-slate-500">No services yet.</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-200 text-left text-xs uppercase text-slate-500">
                  <th className="px-4 py-2">Service</th>
                  <th className="px-4 py-2">Target</th>
                  <th className="px-4 py-2">Status</th>
                  <th className="px-4 py-2 text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((service, index) => (
                  <tr key={service.id} className="border-b border-slate-100 last:border-0">
                    <td className="px-4 py-2">
                      <Link
                        className="font-medium text-slate-900 hover:underline"
                        to="/services/$service"
                        params={{ service: service.slug }}
                      >
                        {service.name}
                      </Link>
                      <span className="ml-2 text-xs uppercase text-slate-400">{service.type}</span>
                    </td>
                    <td className="max-w-xs truncate px-4 py-2 text-slate-600">{serviceTarget(service)}</td>
                    <td className="px-4 py-2">
                      <StateBadge state={service.enabled ? statuses[index]?.data?.state : "paused"} />
                    </td>
                    <td className="px-4 py-2 text-right">
                      {canWrite ? (
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="secondary"
                            onClick={() => toggle.mutate({ name: service.slug, enabled: service.enabled })}
                          >
                            {service.enabled ? "Pause" : "Resume"}
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={() => void navigate({ to: "/services/$service/edit", params: { service: service.slug } })}
                          >
                            Edit
                          </Button>
                          <Button
                            variant="danger"
                            onClick={() => {
                              if (window.confirm(`Delete ${service.name}?`)) remove.mutate(service.slug);
                            }}
                          >
                            Delete
                          </Button>
                        </div>
                      ) : (
                        <span className="text-xs text-slate-400">—</span>
                      )}
                    </td>
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
