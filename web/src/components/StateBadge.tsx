import type { State } from "../lib/types";
import { Badge } from "./ui";

const labels: Record<State, string> = {
  up: "Up",
  down: "Down",
  pending: "Pending",
  paused: "Paused",
};

export function StateBadge({ state }: { state: State | undefined }) {
  if (!state) return <Badge tone="neutral">Unknown</Badge>;
  return <Badge tone={state}>{labels[state]}</Badge>;
}
