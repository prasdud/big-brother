import { cn } from "@/lib/utils";
import type { Heartbeat } from "@/lib/types";

const TOTAL = 30;

function barClass(status: string | undefined): string {
  if (status === "up") return "bg-primary";
  if (status === "down") return "bg-destructive";
  return "bg-border";
}

function slots(beats: Heartbeat[], total = TOTAL): (Heartbeat | null)[] {
  const tail = beats.slice(-total);
  const pad = Math.max(0, total - tail.length);
  return [...Array.from({ length: pad }, () => null), ...tail];
}

/** Miniature uptime bars for a monitor row. */
export function HeartbeatBars({ beats }: { beats: Heartbeat[] }) {
  return (
    <div className="flex shrink-0 items-end gap-[2px]" aria-hidden>
      {slots(beats).map((beat, index) => (
        <span key={index} className={cn("h-6 w-1 rounded-full", barClass(beat?.status))} />
      ))}
    </div>
  );
}

/** Wide uptime timeline for the monitor status card. */
export function Timeline({ beats }: { beats: Heartbeat[] }) {
  return (
    <div className="flex h-10 items-stretch gap-1" aria-hidden>
      {slots(beats).map((beat, index) => (
        <span key={index} className={cn("flex-1 rounded-md", barClass(beat?.status))} />
      ))}
    </div>
  );
}

/** Minimal response-time line chart with a subtle grid. */
export function ResponseGraph({ beats }: { beats: Heartbeat[] }) {
  const values = beats.map((beat) => beat.latency_ms);
  const max = Math.max(1, ...values);
  const count = values.length;
  const W = 100;
  const H = 100;
  const step = count > 1 ? W / (count - 1) : 0;
  const points = values
    .map((value, index) => `${(index * step).toFixed(2)},${(H - (value / max) * H).toFixed(2)}`)
    .join(" ");

  const ticks = [0, 0.25, 0.5, 0.75, 1].map((fraction) => ({
    y: fraction * H,
    label: `${Math.round(max * (1 - fraction))}`,
  }));

  return (
    <div className="flex gap-3">
      <div className="flex flex-col justify-between py-1 text-[10px] tabular-nums text-muted-foreground">
        {ticks.map((tick) => (
          <span key={tick.y}>{tick.label}</span>
        ))}
      </div>
      <div className="h-52 flex-1">
        <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="h-full w-full">
          {ticks.map((tick) => (
            <line
              key={`h-${tick.y}`}
              x1="0"
              x2="100"
              y1={tick.y}
              y2={tick.y}
              className="stroke-border"
              strokeWidth="0.2"
              vectorEffect="non-scaling-stroke"
            />
          ))}
          {[0, 1, 2, 3, 4].map((index) => (
            <line
              key={`v-${index}`}
              x1={index * 25}
              x2={index * 25}
              y1="0"
              y2="100"
              className="stroke-border"
              strokeWidth="0.2"
              vectorEffect="non-scaling-stroke"
            />
          ))}
          {count > 1 ? (
            <polyline
              points={points}
              fill="none"
              className="stroke-primary"
              strokeWidth="1.5"
              strokeLinejoin="round"
              strokeLinecap="round"
              vectorEffect="non-scaling-stroke"
            />
          ) : null}
        </svg>
        {count === 0 ? (
          <p className="mt-2 text-xs text-muted-foreground">No response data yet.</p>
        ) : null}
      </div>
    </div>
  );
}
