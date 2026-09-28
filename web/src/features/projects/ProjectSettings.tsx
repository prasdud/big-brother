import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { projectsApi } from "../../lib/api";
import { errorMessage } from "../../lib/format";
import { useProjects } from "../../lib/project";
import { useSession } from "../../lib/session";
import { ChannelPicker } from "../../components/ChannelPicker";
import { useToast } from "../../components/Toast";
import { Button, Card, CardBody, CardHeader, ErrorText, Field, Input } from "../../components/ui";

export function ProjectSettings() {
  const { project, refresh } = useProjects();
  const { isAdmin, canWrite } = useSession();
  const queryClient = useQueryClient();
  const toast = useToast();
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (project) setName(project.name);
  }, [project]);

  const rename = useMutation({
    mutationFn: () => projectsApi.rename(project!.slug, name),
    onSuccess: () => {
      toast.push("success", "Project renamed");
      refresh();
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
    onError: (err) => setError(errorMessage(err)),
  });

  const remove = useMutation({
    mutationFn: () => projectsApi.remove(project!.slug),
    onSuccess: () => {
      toast.push("success", "Project deleted");
      localStorage.removeItem("bb.project");
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      void navigate({ to: "/" });
    },
    onError: (err) => toast.push("error", errorMessage(err)),
  });

  if (!project) return <p className="text-sm text-slate-600">Select a project first.</p>;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader title="Project" />
        <CardBody className="space-y-4">
          <Field label="Name">
            <Input value={name} disabled={!isAdmin} onChange={(e) => setName(e.target.value)} />
          </Field>
          <ErrorText>{error}</ErrorText>
          {isAdmin ? (
            <div className="flex gap-2">
              <Button
                disabled={rename.isPending || !name}
                onClick={() => {
                  setError("");
                  rename.mutate();
                }}
              >
                Save
              </Button>
              <Button
                variant="danger"
                onClick={() => {
                  if (window.confirm(`Delete ${project.name} and all its services?`)) remove.mutate();
                }}
              >
                Delete project
              </Button>
            </div>
          ) : (
            <p className="text-xs text-slate-500">Only admins can rename or delete projects.</p>
          )}
        </CardBody>
      </Card>

      <Card>
        <CardHeader title="Default alert channel" />
        <CardBody>
          <ChannelPicker project={project.slug} currentChannelID={project.default_channel_id} canWrite={canWrite} />
        </CardBody>
      </Card>
    </div>
  );
}
