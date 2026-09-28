import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import { servicesApi } from "../../lib/api";
import { errorMessage } from "../../lib/format";
import { useProjects } from "../../lib/project";
import { useSession } from "../../lib/session";
import { useToast } from "../../components/Toast";
import { Button, Card, CardBody, CardHeader, ErrorText, Field, Input, Select, Spinner, Textarea } from "../../components/ui";
import type { Service, ServiceInput, ServiceType } from "../../lib/types";

const defaults: ServiceInput = {
  name: "",
  type: "http",
  url: "",
  hostname: "",
  port: 0,
  interval_seconds: 60,
  timeout_seconds: 10,
  failure_threshold: 3,
  template_down: "",
  template_recovered: "",
};

function toInput(service: Service): ServiceInput {
  return {
    name: service.name,
    type: service.type,
    url: service.url,
    hostname: service.hostname,
    port: service.port,
    interval_seconds: service.interval_seconds,
    timeout_seconds: service.timeout_seconds,
    failure_threshold: service.failure_threshold,
    template_down: service.template_down,
    template_recovered: service.template_recovered,
  };
}

function validate(form: ServiceInput): string {
  if (!form.name.trim()) return "Name is required";
  if (form.type === "http" && !form.url.trim()) return "URL is required for HTTP services";
  if ((form.type === "tcp" || form.type === "dns") && !form.hostname.trim()) return "Hostname is required";
  if (form.type === "tcp" && (form.port < 1 || form.port > 65535)) return "Port must be between 1 and 65535";
  if (form.interval_seconds < 1) return "Interval must be positive";
  if (form.timeout_seconds < 1) return "Timeout must be positive";
  if (form.failure_threshold < 1) return "Failure threshold must be positive";
  return "";
}

export function ServiceForm() {
  const params = useParams({ strict: false }) as { service?: string };
  const editing = Boolean(params.service);
  const { project } = useProjects();
  const { canWrite } = useSession();
  const navigate = useNavigate();
  const toast = useToast();
  const queryClient = useQueryClient();
  const slug = project?.slug ?? "";

  const [form, setForm] = useState<ServiceInput>(defaults);
  const [error, setError] = useState("");

  const existing = useQuery({
    queryKey: ["service", slug, params.service],
    queryFn: () => servicesApi.get(slug, params.service as string),
    enabled: editing && Boolean(slug),
  });

  useEffect(() => {
    if (existing.data) setForm(toInput(existing.data));
  }, [existing.data]);

  const save = useMutation({
    mutationFn: () => (editing ? servicesApi.update(slug, params.service as string, form) : servicesApi.create(slug, form)),
    onSuccess: (service) => {
      toast.push("success", editing ? "Service updated" : "Service created");
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
      void navigate({ to: "/services/$service", params: { service: service.slug } });
    },
    onError: (err) => setError(errorMessage(err)),
  });

  if (!project) return <p className="text-sm text-slate-600">Select a project first.</p>;
  if (!canWrite) return <p className="text-sm text-slate-600">You do not have permission to edit services.</p>;
  if (editing && existing.isLoading) return <Spinner />;

  const set = <K extends keyof ServiceInput>(key: K, value: ServiceInput[K]) =>
    setForm((current) => ({ ...current, [key]: value }));

  return (
    <Card className="mx-auto max-w-2xl">
      <CardHeader title={editing ? "Edit service" : "New service"} />
      <CardBody>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            const problem = validate(form);
            setError(problem);
            if (!problem) save.mutate();
          }}
        >
          <div className="grid grid-cols-2 gap-4">
            <Field label="Name">
              <Input value={form.name} onChange={(e) => set("name", e.target.value)} />
            </Field>
            <Field label="Type">
              <Select value={form.type} onChange={(e) => set("type", e.target.value as ServiceType)} disabled={editing}>
                <option value="http">HTTP</option>
                <option value="tcp">TCP</option>
                <option value="dns">DNS</option>
              </Select>
            </Field>
          </div>

          {form.type === "http" ? (
            <Field label="URL" hint="Checked with a GET request; 2xx and 3xx count as up.">
              <Input value={form.url} onChange={(e) => set("url", e.target.value)} placeholder="https://example.com/health" />
            </Field>
          ) : (
            <div className="grid grid-cols-3 gap-4">
              <div className="col-span-2">
                <Field label="Hostname">
                  <Input value={form.hostname} onChange={(e) => set("hostname", e.target.value)} />
                </Field>
              </div>
              {form.type === "tcp" ? (
                <Field label="Port">
                  <Input
                    type="number"
                    value={form.port}
                    onChange={(e) => set("port", Number(e.target.value))}
                  />
                </Field>
              ) : null}
            </div>
          )}

          <div className="grid grid-cols-3 gap-4">
            <Field label="Interval (s)">
              <Input
                type="number"
                value={form.interval_seconds}
                onChange={(e) => set("interval_seconds", Number(e.target.value))}
              />
            </Field>
            <Field label="Timeout (s)">
              <Input
                type="number"
                value={form.timeout_seconds}
                onChange={(e) => set("timeout_seconds", Number(e.target.value))}
              />
            </Field>
            <Field label="Failure threshold">
              <Input
                type="number"
                value={form.failure_threshold}
                onChange={(e) => set("failure_threshold", Number(e.target.value))}
              />
            </Field>
          </div>

          <details className="rounded-md border border-slate-200 p-3">
            <summary className="cursor-pointer text-sm font-medium text-slate-700">
              Template overrides (optional)
            </summary>
            <div className="mt-3 space-y-3">
              <Field label="Down template" hint="Leave blank to use the project or built-in template.">
                <Textarea
                  rows={3}
                  value={form.template_down}
                  onChange={(e) => set("template_down", e.target.value)}
                />
              </Field>
              <Field label="Recovered template">
                <Textarea
                  rows={3}
                  value={form.template_recovered}
                  onChange={(e) => set("template_recovered", e.target.value)}
                />
              </Field>
            </div>
          </details>

          <ErrorText>{error}</ErrorText>
          <div className="flex gap-2">
            <Button type="submit" disabled={save.isPending}>
              {editing ? "Save changes" : "Create service"}
            </Button>
            <Button type="button" variant="secondary" onClick={() => void navigate({ to: "/" })}>
              Cancel
            </Button>
          </div>
        </form>
      </CardBody>
    </Card>
  );
}
