import { useQuery } from "@tanstack/react-query";
import { slackApi } from "../../lib/api";
import { useSession } from "../../lib/session";
import { Button, Card, CardBody, CardHeader, Spinner } from "../../components/ui";

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
        <CardHeader title="Slack workspace" />
        <CardBody className="space-y-3">
          {status.isLoading ? (
            <Spinner />
          ) : connected ? (
            <p className="text-sm text-slate-700">
              Connected to <span className="font-medium">{status.data?.team_name}</span>.
            </p>
          ) : (
            <p className="text-sm text-slate-600">No Slack workspace connected.</p>
          )}
          {isAdmin ? (
            <a href="/api/v1/slack/install">
              <Button>{connected ? "Reconnect Slack" : "Connect Slack"}</Button>
            </a>
          ) : (
            <p className="text-xs text-slate-500">Only admins can connect Slack.</p>
          )}
        </CardBody>
      </Card>

      {connected ? (
        <Card>
          <CardHeader title="Available channels" />
          <CardBody className="p-0">
            {channels.isLoading ? (
              <div className="p-4">
                <Spinner />
              </div>
            ) : (
              <ul className="divide-y divide-slate-100">
                {(channels.data ?? []).map((channel) => (
                  <li key={channel.id} className="px-4 py-2 text-sm text-slate-700">
                    #{channel.name}
                  </li>
                ))}
              </ul>
            )}
          </CardBody>
        </Card>
      ) : null}
    </div>
  );
}
