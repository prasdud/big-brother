import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import { toast } from "sonner";
import { servicesApi } from "@/lib/api";
import { errorMessage } from "@/lib/format";
import { useProjects } from "@/lib/project";
import { useSession } from "@/lib/session";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import type { Service, ServiceInput, ServiceType } from "@/lib/types";

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
  tags: [],
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
    tags: service.tags ?? [],
  };
}

function parseTags(input: string): string[] {
  return input
    .split(",")
    .map((tag) => tag.trim())
    .filter(Boolean);
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
  const queryClient = useQueryClient();
  const slug = project?.slug ?? "";

  const [form, setForm] = useState<ServiceInput>(defaults);
  const [tagsInput, setTagsInput] = useState("");
  const [error, setError] = useState("");

  const existing = useQuery({
    queryKey: ["service", slug, params.service],
    queryFn: () => servicesApi.get(slug, params.service as string),
    enabled: editing && Boolean(slug),
  });

  useEffect(() => {
    if (existing.data) {
      setForm(toInput(existing.data));
      setTagsInput((existing.data.tags ?? []).join(", "));
    }
  }, [existing.data]);

  const save = useMutation({
    mutationFn: () => {
      const input = { ...form, tags: parseTags(tagsInput) };
      return editing
        ? servicesApi.update(slug, params.service as string, input)
        : servicesApi.create(slug, input);
    },
    onSuccess: () => {
      toast.success(editing ? "Service updated" : "Service created");
      void queryClient.invalidateQueries({ queryKey: ["services", slug] });
      void queryClient.invalidateQueries({ queryKey: ["monitors", slug] });
      void navigate({ to: "/services" });
    },
    onError: (err) => setError(errorMessage(err)),
  });

  if (!project) return <p className="text-sm text-muted-foreground">Select a project first.</p>;
  if (!canWrite) return <p className="text-sm text-muted-foreground">You do not have permission to edit services.</p>;
  if (editing && existing.isLoading) return <Skeleton className="h-96 w-full" />;

  const set = <K extends keyof ServiceInput>(key: K, value: ServiceInput[K]) =>
    setForm((current) => ({ ...current, [key]: value }));

  return (
    <Card className="mx-auto w-full max-w-2xl">
      <CardHeader>
        <CardTitle>{editing ? "Edit service" : "New service"}</CardTitle>
        <CardDescription>One service equals one check configuration.</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          className="space-y-6"
          onSubmit={(event) => {
            event.preventDefault();
            const problem = validate(form);
            setError(problem);
            if (!problem) save.mutate();
          }}
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={form.name} onChange={(e) => set("name", e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Type</Label>
              <Select
                value={form.type}
                onValueChange={(value) => set("type", value as ServiceType)}
                disabled={editing}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="http">HTTP</SelectItem>
                  <SelectItem value="tcp">TCP</SelectItem>
                  <SelectItem value="dns">DNS</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {form.type === "http" ? (
            <div className="space-y-2">
              <Label htmlFor="url">URL</Label>
              <Input
                id="url"
                value={form.url}
                onChange={(e) => set("url", e.target.value)}
                placeholder="https://example.com/health"
              />
              <p className="text-label text-muted-foreground">GET request; 2xx and 3xx count as up.</p>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="hostname">Hostname</Label>
                <Input
                  id="hostname"
                  value={form.hostname}
                  onChange={(e) => set("hostname", e.target.value)}
                />
              </div>
              {form.type === "tcp" ? (
                <div className="space-y-2">
                  <Label htmlFor="port">Port</Label>
                  <Input
                    id="port"
                    type="number"
                    value={form.port}
                    onChange={(e) => set("port", Number(e.target.value))}
                  />
                </div>
              ) : null}
            </div>
          )}

          <div className="grid gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="interval">Interval (s)</Label>
              <Input
                id="interval"
                type="number"
                value={form.interval_seconds}
                onChange={(e) => set("interval_seconds", Number(e.target.value))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="timeout">Timeout (s)</Label>
              <Input
                id="timeout"
                type="number"
                value={form.timeout_seconds}
                onChange={(e) => set("timeout_seconds", Number(e.target.value))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="threshold">Failure threshold</Label>
              <Input
                id="threshold"
                type="number"
                value={form.failure_threshold}
                onChange={(e) => set("failure_threshold", Number(e.target.value))}
              />
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <div>
              <p className="text-sm font-medium">Template overrides</p>
              <p className="text-label text-muted-foreground">
                Leave blank to use the project or built-in template.
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="template-down">Down template</Label>
              <Textarea
                id="template-down"
                rows={3}
                value={form.template_down}
                onChange={(e) => set("template_down", e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="template-recovered">Recovered template</Label>
              <Textarea
                id="template-recovered"
                rows={3}
                value={form.template_recovered}
                onChange={(e) => set("template_recovered", e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="tags">Tags</Label>
            <Input
              id="tags"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              placeholder="pay, api"
            />
            <p className="text-label text-muted-foreground">Comma-separated.</p>
          </div>

          {error ? <p className="text-sm text-destructive">{error}</p> : null}

          <div className="flex gap-2">
            <Button type="submit" disabled={save.isPending}>
              {editing ? "Save changes" : "Create service"}
            </Button>
            <Button type="button" variant="outline" onClick={() => void navigate({ to: "/services" })}>
              Cancel
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
