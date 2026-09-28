import { useState } from "react";
import { useMutation, useQueries, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { FolderKanban, Plus } from "lucide-react";
import { toast } from "sonner";
import { projectsApi, servicesApi } from "@/lib/api";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import type { Project } from "@/lib/types";

function CreateProjectDialog() {
  const queryClient = useQueryClient();
  const { select } = useProjects();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");

  const create = useMutation({
    mutationFn: () => projectsApi.create(name),
    onSuccess: (project) => {
      toast.success("Project created");
      select(project.slug);
      setName("");
      setOpen(false);
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      void navigate({ to: "/services" });
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus />
          New project
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New project</DialogTitle>
          <DialogDescription>A project groups the services you monitor together.</DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault();
            if (name) create.mutate();
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="new-project-name">Project name</Label>
            <Input
              id="new-project-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Payments"
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button type="submit" disabled={create.isPending || !name}>
              Create project
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function ProjectCard({ project, count }: { project: Project; count: number | undefined }) {
  const { select } = useProjects();
  const navigate = useNavigate();

  return (
    <Card
      role="button"
      tabIndex={0}
      className="cursor-pointer transition-colors hover:border-primary/60"
      onClick={() => {
        select(project.slug);
        void navigate({ to: "/services" });
      }}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          select(project.slug);
          void navigate({ to: "/services" });
        }
      }}
    >
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FolderKanban className="size-4 text-muted-foreground" />
          {project.name}
        </CardTitle>
        <CardDescription>/{project.slug}</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">
          {count === undefined ? "…" : `${count} service${count === 1 ? "" : "s"}`}
        </p>
      </CardContent>
    </Card>
  );
}

export function Projects() {
  const { projects, loading } = useProjects();
  const { isAdmin } = useSession();

  const counts = useQueries({
    queries: projects.map((project) => ({
      queryKey: ["services", project.slug],
      queryFn: () => servicesApi.list(project.slug),
    })),
  });

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h1 className="text-h1">Projects</h1>
          <p className="text-sm text-muted-foreground">
            {loading ? "Loading…" : `${projects.length} project${projects.length === 1 ? "" : "s"}`}
          </p>
        </div>
        {isAdmin ? <CreateProjectDialog /> : null}
      </div>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-32 w-full" />
        </div>
      ) : projects.length === 0 ? (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            No projects yet. {isAdmin ? "Create one to get started." : "Ask an admin to create one."}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((project, index) => (
            <ProjectCard key={project.id} project={project} count={counts[index]?.data?.length} />
          ))}
        </div>
      )}
    </div>
  );
}
