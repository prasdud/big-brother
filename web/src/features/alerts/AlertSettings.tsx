import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { alertsApi } from "@/lib/api";
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
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";

const variables = ["service.name", "service.url", "project.name", "status", "duration", "error"];

interface EditorProps {
  project: string;
  trigger: "down" | "recovered";
  title: string;
  description: string;
  body: string;
  setBody: (value: string) => void;
  channelID: string;
  canWrite: boolean;
}

function TemplateEditor({
  project,
  trigger,
  title,
  description,
  body,
  setBody,
  channelID,
  canWrite,
}: EditorProps) {
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
    onSuccess: () => toast.success("Template saved"),
    onError: (error) => toast.error(errorMessage(error)),
  });

  const test = useMutation({
    mutationFn: () => alertsApi.testSend(project, trigger, body, channelID),
    onSuccess: () => toast.success("Test alert sent"),
    onError: (error) => toast.error(errorMessage(error)),
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
        {canWrite ? (
          <div className="ml-auto flex gap-2">
            <Button variant="outline" size="sm" onClick={() => test.mutate()} disabled={test.isPending}>
              Test send
            </Button>
            <Button size="sm" onClick={() => save.mutate()} disabled={save.isPending || !body.trim()}>
              Save
            </Button>
          </div>
        ) : null}
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex flex-wrap gap-1">
          {variables.map((variable) => (
            <Button
              key={variable}
              type="button"
              variant="outline"
              size="xs"
              disabled={!canWrite}
              onClick={() => setBody(`${body}{{${variable}}}`)}
            >
              {`{{${variable}}}`}
            </Button>
          ))}
        </div>
        <Textarea rows={6} value={body} disabled={!canWrite} onChange={(e) => setBody(e.target.value)} />
        <div>
          <p className="mb-1 text-label uppercase text-muted-foreground">Preview</p>
          <pre className="min-h-[4rem] whitespace-pre-wrap rounded-lg border border-border bg-muted/40 p-3 text-sm">
            {preview || "—"}
          </pre>
        </div>
      </CardContent>
    </Card>
  );
}

export function AlertSettings() {
  const { project } = useProjects();
  const { canWrite } = useSession();
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

  if (!project) return <p className="text-sm text-muted-foreground">Select a project first.</p>;
  if (templates.isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-h1">Alert templates</h1>
        <p className="text-sm text-muted-foreground">Messages sent to Slack on state changes.</p>
      </div>
      <TemplateEditor
        project={project.slug}
        trigger="down"
        title="Down alert"
        description="Sent when a service is confirmed down."
        body={down}
        setBody={setDown}
        channelID={project.default_channel_id}
        canWrite={canWrite}
      />
      <TemplateEditor
        project={project.slug}
        trigger="recovered"
        title="Recovery alert"
        description="Sent when a service recovers."
        body={recovered}
        setBody={setRecovered}
        channelID={project.default_channel_id}
        canWrite={canWrite}
      />
    </div>
  );
}
