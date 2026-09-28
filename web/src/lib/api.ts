import type {
  AlertTemplate,
  ChannelOption,
  ChannelView,
  Check,
  Delivery,
  Project,
  Service,
  ServiceInput,
  SlackStatus,
  Status,
  Uptime,
  User,
} from "./types";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

function csrfToken(): string | null {
  const match = document.cookie.split("; ").find((row) => row.startsWith("bb_csrf="));
  return match ? decodeURIComponent(match.split("=").slice(1).join("=")) : null;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (method !== "GET") {
    const token = csrfToken();
    if (token) headers["X-CSRF-Token"] = token;
  }

  const res = await fetch(path, {
    method,
    headers,
    credentials: "same-origin",
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  if (res.status === 204) return undefined as T;

  const data = await res.json().catch(() => null);
  if (!res.ok) {
    const message = (data as { error?: { message?: string } } | null)?.error?.message ?? res.statusText;
    throw new ApiError(res.status, message);
  }
  return data as T;
}

export const sessionApi = {
  me: () => request<User>("GET", "/auth/me"),
  logout: () => request<void>("POST", "/auth/logout"),
};

export const projectsApi = {
  list: () => request<Project[]>("GET", "/api/v1/projects"),
  create: (name: string) => request<Project>("POST", "/api/v1/projects", { name }),
  rename: (slug: string, name: string) => request<Project>("PATCH", `/api/v1/projects/${slug}`, { name }),
  remove: (slug: string) => request<void>("DELETE", `/api/v1/projects/${slug}`),
};

export const servicesApi = {
  list: (project: string) => request<Service[]>("GET", `/api/v1/projects/${project}/services`),
  get: (project: string, service: string) =>
    request<Service>("GET", `/api/v1/projects/${project}/services/${service}`),
  create: (project: string, input: ServiceInput) =>
    request<Service>("POST", `/api/v1/projects/${project}/services`, input),
  update: (project: string, service: string, input: ServiceInput) =>
    request<Service>("PATCH", `/api/v1/projects/${project}/services/${service}`, input),
  remove: (project: string, service: string) =>
    request<void>("DELETE", `/api/v1/projects/${project}/services/${service}`),
  pause: (project: string, service: string) =>
    request<Service>("POST", `/api/v1/projects/${project}/services/${service}/pause`),
  resume: (project: string, service: string) =>
    request<Service>("POST", `/api/v1/projects/${project}/services/${service}/resume`),
  status: (project: string, service: string) =>
    request<Status>("GET", `/api/v1/projects/${project}/services/${service}/status`),
  checks: (project: string, service: string, limit = 50) =>
    request<Check[]>("GET", `/api/v1/projects/${project}/services/${service}/checks?limit=${limit}`),
  uptime: (project: string, service: string, window = "24h") =>
    request<Uptime>("GET", `/api/v1/projects/${project}/services/${service}/uptime?window=${window}`),
  setChannel: (project: string, service: string, slackChannelID: string, name: string) =>
    request<ChannelView>("PUT", `/api/v1/projects/${project}/services/${service}/channel`, {
      slack_channel_id: slackChannelID,
      name,
    }),
  clearChannel: (project: string, service: string) =>
    request<void>("DELETE", `/api/v1/projects/${project}/services/${service}/channel`),
};

export const alertsApi = {
  templates: (project: string) =>
    request<AlertTemplate[]>("GET", `/api/v1/projects/${project}/alert-templates`),
  saveTemplate: (project: string, trigger: string, body: string) =>
    request<AlertTemplate>("PUT", `/api/v1/projects/${project}/alert-templates/${trigger}`, { body }),
  preview: (project: string, trigger: string, body: string) =>
    request<{ rendered: string }>("POST", `/api/v1/projects/${project}/alert-templates/preview`, {
      trigger,
      body,
    }),
  testSend: (project: string, trigger: string, body: string, channelID: string) =>
    request<{ sent: boolean }>("POST", `/api/v1/projects/${project}/alert-templates/test-send`, {
      trigger,
      body,
      channel_id: channelID,
    }),
  deliveries: (project: string) =>
    request<Delivery[]>("GET", `/api/v1/projects/${project}/deliveries`),
};

export const slackApi = {
  status: () => request<SlackStatus>("GET", "/api/v1/slack"),
  channels: () => request<ChannelOption[]>("GET", "/api/v1/slack/channels"),
  setProjectChannel: (project: string, slackChannelID: string, name: string) =>
    request<ChannelView>("PUT", `/api/v1/projects/${project}/channel`, {
      slack_channel_id: slackChannelID,
      name,
    }),
  clearProjectChannel: (project: string) =>
    request<void>("DELETE", `/api/v1/projects/${project}/channel`),
  projectChannels: (project: string) =>
    request<ChannelView[]>("GET", `/api/v1/projects/${project}/channels`),
};

export const usersApi = {
  list: () => request<User[]>("GET", "/api/v1/users"),
  create: (email: string, name: string, role: string) =>
    request<User>("POST", "/api/v1/users", { email, name, role }),
  update: (id: string, patch: { role?: string; disabled?: boolean }) =>
    request<User>("PATCH", `/api/v1/users/${id}`, patch),
  remove: (id: string) => request<void>("DELETE", `/api/v1/users/${id}`),
};
