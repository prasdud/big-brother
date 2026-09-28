import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { projectsApi } from "@/lib/api";
import { errorMessage } from "@/lib/format";
import { useProjects } from "@/lib/project";
import { useSession } from "@/lib/session";
import { ChannelPicker } from "@/components/ChannelPicker";
import { ConfirmDialog } from "@/components/confirm-dialog";
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

export function ProjectSettings() {
  const { project, refresh } = useProjects();
  const { isAdmin, canWrite } = useSession();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [confirming, setConfirming] = useState(false);

  useEffect(() => {
    if (project) setName(project.name);
  }, [project]);

  const rename = useMutation({
    mutationFn: () => projectsApi.rename(project!.slug, name),
    onSuccess: () => {
      toast.success("Project renamed");
      refresh();
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
    onError: (err) => toast.error(errorMessage(err)),
  });

  const remove = useMutation({
    mutationFn: () => projectsApi.remove(project!.slug),
    onSuccess: () => {
      toast.success("Project deleted");
      localStorage.removeItem("bb.project");
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      void navigate({ to: "/" });
    },
    onError: (err) => toast.error(errorMessage(err)),
  });

  if (!project) return <p className="text-sm text-muted-foreground">Select a project first.</p>;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Project</CardTitle>
          <CardDescription>Rename or delete this project.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="max-w-sm space-y-2">
            <Label htmlFor="project-name">Name</Label>
            <Input
              id="project-name"
              value={name}
              disabled={!isAdmin}
              onChange={(event) => setName(event.target.value)}
            />
          </div>
          {isAdmin ? (
            <div className="flex gap-2">
              <Button
                disabled={rename.isPending || !name}
                onClick={() => rename.mutate()}
              >
                Save
              </Button>
              <Button variant="destructive" onClick={() => setConfirming(true)}>
                Delete project
              </Button>
            </div>
          ) : (
            <p className="text-label text-muted-foreground">
              Only admins can rename or delete projects.
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Default alert channel</CardTitle>
          <CardDescription>Used when a service has no channel override.</CardDescription>
        </CardHeader>
        <CardContent>
          <ChannelPicker
            project={project.slug}
            currentChannelID={project.default_channel_id}
            canWrite={canWrite}
          />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={confirming}
        onOpenChange={setConfirming}
        title={`Delete ${project.name}?`}
        description="This removes the project and all its services."
        confirmLabel="Delete project"
        pending={remove.isPending}
        onConfirm={() => remove.mutate()}
      />
    </div>
  );
}
