import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { toast } from "sonner";
import { servicesApi, slackApi } from "@/lib/api";
import { errorMessage } from "@/lib/format";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface ChannelPickerProps {
  project: string;
  service?: string;
  currentChannelID: string;
  canWrite: boolean;
}

export function ChannelPicker({ project, service, currentChannelID, canWrite }: ChannelPickerProps) {
  const queryClient = useQueryClient();

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
      toast.success("Channel saved");
      invalidate();
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  const clear = useMutation({
    mutationFn: () =>
      service ? servicesApi.clearChannel(project, service) : slackApi.clearProjectChannel(project),
    onSuccess: () => {
      toast.success("Channel cleared");
      invalidate();
    },
    onError: (error) => toast.error(errorMessage(error)),
  });

  if (!connected) {
    return (
      <p className="text-sm text-muted-foreground">
        Slack is not connected.{" "}
        <Link className="text-primary underline-offset-4 hover:underline" to="/integrations">
          Connect Slack
        </Link>{" "}
        to choose a channel.
      </p>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Select
        value={currentSlackID || "none"}
        disabled={!canWrite}
        onValueChange={(value) => {
          if (value === "none") {
            if (currentSlackID) clear.mutate();
            return;
          }
          const channel = channels.data?.find((c) => c.id === value);
          if (channel) save.mutate({ id: channel.id, name: channel.name });
        }}
      >
        <SelectTrigger className="w-64">
          <SelectValue placeholder="No channel" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="none">No channel</SelectItem>
          {channels.data?.map((channel) => (
            <SelectItem key={channel.id} value={channel.id}>
              #{channel.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {save.isPending || clear.isPending ? (
        <span className="text-xs text-muted-foreground">Saving…</span>
      ) : null}
    </div>
  );
}
