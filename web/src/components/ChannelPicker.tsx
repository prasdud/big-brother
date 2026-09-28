import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { servicesApi, slackApi } from "../lib/api";
import { errorMessage } from "../lib/format";
import { Button, Select } from "./ui";
import { useToast } from "./Toast";

interface ChannelPickerProps {
  project: string;
  service?: string;
  currentChannelID: string;
  canWrite: boolean;
}

export function ChannelPicker({ project, service, currentChannelID, canWrite }: ChannelPickerProps) {
  const queryClient = useQueryClient();
  const toast = useToast();

  const status = useQuery({ queryKey: ["slack"], queryFn: slackApi.status });
  const connected = status.data?.connected ?? false;
  const channels = useQuery({
    queryKey: ["slack", "channels"],
    queryFn: slackApi.channels,
    enabled: connected,
  });
  const projectChannels = useQuery({
    queryKey: ["project-channels", project],
    queryFn: () => slackApi.projectChannels(project),
    enabled: connected,
  });

  const currentSlackID =
    projectChannels.data?.find((c) => c.channel_id === currentChannelID)?.slack_channel_id ?? "";

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ["slack"] });
    void queryClient.invalidateQueries({ queryKey: ["project-channels", project] });
    void queryClient.invalidateQueries({ queryKey: ["projects"] });
    void queryClient.invalidateQueries({ queryKey: ["service", project, service] });
  };

  const save = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      service
        ? servicesApi.setChannel(project, service, id, name)
        : slackApi.setProjectChannel(project, id, name),
    onSuccess: () => {
      toast.push("success", "Channel saved");
      invalidate();
    },
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  const clear = useMutation({
    mutationFn: () =>
      service ? servicesApi.clearChannel(project, service) : slackApi.clearProjectChannel(project),
    onSuccess: () => {
      toast.push("success", "Channel cleared");
      invalidate();
    },
    onError: (error) => toast.push("error", errorMessage(error)),
  });

  if (!connected) {
    return (
      <p className="text-sm text-slate-600">
        Slack is not connected. <Link className="underline" to="/slack">Connect Slack</Link> to choose a channel.
      </p>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Select
        value={currentSlackID}
        disabled={!canWrite}
        onChange={(e) => {
          const channel = channels.data?.find((c) => c.id === e.target.value);
          if (channel) save.mutate({ id: channel.id, name: channel.name });
        }}
      >
        <option value="">No channel</option>
        {channels.data?.map((channel) => (
          <option key={channel.id} value={channel.id}>
            #{channel.name}
          </option>
        ))}
      </Select>
      {currentSlackID && canWrite ? (
        <Button variant="secondary" onClick={() => clear.mutate()}>
          Clear
        </Button>
      ) : null}
    </div>
  );
}
