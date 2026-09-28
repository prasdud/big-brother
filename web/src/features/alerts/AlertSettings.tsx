import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { alertsApi } from "../../lib/api";
import { errorMessage } from "../../lib/format";
import { useProjects } from "../../lib/project";
import { useSession } from "../../lib/session";
import { useToast } from "../../components/Toast";
import { Button, Card, CardBody, CardHeader, Spinner, Textarea } from "../../components/ui";

const variables = ["service.name", "service.url", "project.name", "status", "duration", "error"];

interface EditorProps {
  project: string;
  trigger: "down" | "recovered";
  title: string;
  body: string;
  setBody: (value: string) => void;
  channelID: string;
  canWrite: boolean;
}

function TemplateEditor({ project, trigger, title, body, setBody, channelID, canWrite }: EditorProps) {
  const toast = useToast();
  const [preview, setPreview] = useState("");

  useEffect(() => {
    if (!body.trim()) {
      setPreview("");
      return;
    }
    const handle = setTimeout(() => {
      alertsApi
        .preview(project, trigger, body)
        .then((result) => setPreview(result.rendered))
        .catch(() => setPreview(""));
    }, 400);
    return () => clearTimeout(handle);
  }, [project, trigger, body]);

  const save = useMutation({
    mutationFn: () => alertsApi.saveTemplate(project, trigger, body),
    onSuccess: () => toast.push("success", "Template saved"),
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  const test = useMutation({
    mutationFn: () => alertsApi.testSend(project, trigger, body, channelID),
    onSuccess: () => toast.push("success", "Test alert sent"),
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  const insert = (variable: string) => setBody(`${body}{{${variable}}}`);

  return (
    <Card>
      <CardHeader
        title={title}
        actions={
          canWrite ? (
            <div className="flex gap-2">
              <Button variant="secondary" onClick={() => test.mutate()} disabled={test.isPending}>
                Test send
              </Button>
              <Button onClick={() => save.mutate()} disabled={save.isPending || !body.trim()}>
                Save
              </Button>
            </div>
          ) : null
        }
      />
      <CardBody className="space-y-3">
        <div className="flex flex-wrap gap-1">
          {variables.map((variable) => (
            <button
              key={variable}
              type="button"
              disabled={!canWrite}
              onClick={() => insert(variable)}
              className="rounded bg-slate-100 px-2 py-0.5 text-xs text-slate-700 hover:bg-slate-200 disabled:opacity-50"
            >
              {`{{${variable}}}`}
            </button>
          ))}
        </div>
        <Textarea rows={6} value={body} disabled={!canWrite} onChange={(e) => setBody(e.target.value)} />
        <div>
          <p className="mb-1 text-xs font-medium uppercase text-slate-500">Preview</p>
          <pre className="whitespace-pre-wrap rounded-md bg-slate-50 p-3 text-sm text-slate-700">
            {preview || "—"}
          </pre>
        </div>
      </CardBody>
    </Card>
  );
}

export function AlertSettings() {
  const { project } = useProjects();
  const { canWrite } = useSession();
  const queryClient = useQueryClient();
  const [down, setDown] = useState("");
  const [recovered, setRecovered] = useState("");

  const templates = useQuery({
    queryKey: ["templates", project?.slug],
    queryFn: () => alertsApi.templates(project!.slug),
    enabled: Boolean(project),
  });

  useEffect(() => {
    if (!templates.data) return;
    for (const template of templates.data) {
      if (template.trigger === "down") setDown(template.body);
      if (template.trigger === "recovered") setRecovered(template.body);
    }
  }, [templates.data]);

  useEffect(() => {
    if (project) void queryClient.invalidateQueries({ queryKey: ["project-channels", project.slug] });
  }, [project, queryClient]);

  if (!project) return <p className="text-sm text-slate-600">Select a project first.</p>;
  if (templates.isLoading) return <Spinner />;

  return (
    <div className="space-y-4">
      <h1 className="text-lg font-semibold text-slate-900">Alert templates</h1>
      <TemplateEditor
        project={project.slug}
        trigger="down"
        title="Down alert"
        body={down}
        setBody={setDown}
        channelID={project.default_channel_id}
        canWrite={canWrite}
      />
      <TemplateEditor
        project={project.slug}
        trigger="recovered"
        title="Recovery alert"
        body={recovered}
        setBody={setRecovered}
        channelID={project.default_channel_id}
        canWrite={canWrite}
      />
    </div>
  );
}
