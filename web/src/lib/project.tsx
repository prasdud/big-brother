import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { projectsApi } from "./api";
import type { Project } from "./types";

const STORAGE_KEY = "bb.project";

interface ProjectValue {
  projects: Project[];
  project: Project | null;
  loading: boolean;
  select: (slug: string) => void;
  refresh: () => void;
}

const ProjectContext = createContext<ProjectValue | null>(null);

export function useProjects(): ProjectValue {
  const value = useContext(ProjectContext);
  if (!value) throw new Error("useProjects used outside ProjectProvider");
  return value;
}

export function ProjectProvider({ children }: { children: ReactNode }) {
  const query = useQuery({ queryKey: ["projects"], queryFn: projectsApi.list });
  const [slug, setSlug] = useState<string | null>(() => localStorage.getItem(STORAGE_KEY));

  const projects = useMemo(() => query.data ?? [], [query.data]);

  const project = useMemo(() => {
    if (projects.length === 0) return null;
    return projects.find((p) => p.slug === slug) ?? projects[0];
  }, [projects, slug]);

  const select = useCallback((next: string) => {
    localStorage.setItem(STORAGE_KEY, next);
    setSlug(next);
  }, []);

  const refresh = useCallback(() => {
    void query.refetch();
  }, [query]);

  return (
    <ProjectContext.Provider
      value={{ projects, project, loading: query.isLoading, select, refresh }}
    >
      {children}
    </ProjectContext.Provider>
  );
}
