import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackApi } from "@/lib/api";
import { errorMessage } from "@/lib/format";
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

export function Integrations() {
  const { isAdmin } = useSession();
  const [testing, setTesting] = useState(false);

  const status = useQuery({ queryKey: ["slack"], queryFn: slackApi.status });
  const connected = status.data?.connected ?? false;
  const channels = useQuery({
    queryKey: ["slack", "channels"],
    queryFn: slackApi.channels,
    enabled: connected,
  });

  const testConnection = async () => {
    setTesting(true);
    try {
      const list = await slackApi.channels();
      toast.success(`Slack connection OK · ${list.length} channels visible`);
    } catch (error) {
      toast.error(errorMessage(error));
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-h1">Integrations</h1>
        <p className="text-sm text-muted-foreground">Connect external services for alerts.</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            Slack
            {connected ? (
              <Badge variant="outline" className="border-primary/40 bg-primary/10 text-primary">
                connected
              </Badge>
            ) : (
              <Badge variant="outline" className="text-muted-foreground">
                not connected
              </Badge>
            )}
          </CardTitle>
          <CardDescription>Alerts are posted by the big-brother bot.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {status.isLoading ? (
            <Skeleton className="h-9 w-48" />
          ) : (
            <p className="text-sm text-muted-foreground">
              {connected
                ? `Connected to ${status.data?.team_name}.`
                : "No Slack workspace connected yet."}
            </p>
          )}

          <div className="flex flex-wrap gap-2">
            {isAdmin ? (
              <Button asChild>
                <a href="/api/v1/slack/install">{connected ? "Reconnect Slack" : "Connect Slack"}</a>
              </Button>
            ) : null}
            <Button
              variant="outline"
              onClick={testConnection}
              disabled={testing || !connected}
            >
              {testing ? "Testing…" : "Test connection"}
            </Button>
          </div>
          {!isAdmin ? (
            <p className="text-label text-muted-foreground">Only admins can connect Slack.</p>
          ) : null}
        </CardContent>
      </Card>

      {connected ? (
        <Card>
          <CardHeader>
            <CardTitle>Available channels</CardTitle>
            <CardDescription>
              The bot must be invited to a channel before it can post.
            </CardDescription>
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
