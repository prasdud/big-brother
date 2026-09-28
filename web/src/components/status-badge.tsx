import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { State } from "@/lib/types";

const styles: Record<string, { label: string; className: string; dot: string }> = {
  up: { label: "Up", className: "border-primary/40 bg-primary/10 text-primary", dot: "bg-primary" },
  down: {
    label: "Down",
    className: "border-destructive/40 bg-destructive/10 text-destructive",
    dot: "bg-destructive",
  },
  pending: {
    label: "Pending",
    className: "border-border bg-muted text-muted-foreground",
    dot: "bg-muted-foreground",
  },
  paused: {
    label: "Paused",
    className: "border-border bg-muted text-muted-foreground",
    dot: "bg-muted-foreground",
  },
  unknown: {
    label: "Unknown",
    className: "border-border bg-muted text-muted-foreground",
    dot: "bg-muted-foreground",
  },
};

export function StatusBadge({ state }: { state: State | undefined }) {
  const style = styles[state ?? "unknown"] ?? styles.unknown;
  return (
    <Badge variant="outline" className={cn("gap-1.5 font-normal", style.className)}>
      <span className={cn("size-1.5 rounded-full", style.dot)} />
      {style.label}
    </Badge>
  );
}
