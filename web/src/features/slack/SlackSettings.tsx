import { useQuery } from "@tanstack/react-query";
import { slackApi } from "@/lib/api";
import { useSession } from "@/lib/session";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function SlackSettings() {
  const { isAdmin } = useSession();
  const status = useQuery({ queryKey: ["slack"], queryFn: slackApi.status });
  const connected = status.data?.connected ?? false;
  const channels = useQuery({
    queryKey: ["slack", "channels"],
    queryFn: slackApi.channels,
    enabled: connected,
  });

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            Slack workspace
            {connected ? <Badge variant="outline" className="text-primary">connected</Badge> : null}
          </CardTitle>
          <CardDescription>Alerts are posted by the big-brother bot.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {status.isLoading ? (
            <Skeleton className="h-9 w-48" />
          ) : connected ? (
            <p className="text-sm text-muted-foreground">
              Connected to <span className="text-foreground">{status.data?.team_name}</span>.
            </p>
          ) : (
            <p className="text-sm text-muted-foreground">No Slack workspace connected.</p>
          )}
          {isAdmin ? (
            <Button asChild>
              <a href="/api/v1/slack/install">{connected ? "Reconnect Slack" : "Connect Slack"}</a>
            </Button>
          ) : (
            <p className="text-label text-muted-foreground">Only admins can connect Slack.</p>
          )}
        </CardContent>
      </Card>

      {connected ? (
        <Card>
          <CardHeader>
            <CardTitle>Available channels</CardTitle>
            <CardDescription>The bot must be invited to a channel before it can post.</CardDescription>
          </CardHeader>
          <CardContent className="px-0">
            {channels.isLoading ? (
              <div className="space-y-2 px-6">
                <Skeleton className="h-8 w-full" />
              </div>
            ) : (
              <ul className="divide-y divide-border">
                {(channels.data ?? []).map((channel) => (
                  <li key={channel.id} className="px-6 py-2 text-sm text-muted-foreground">
                    #{channel.name}
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
