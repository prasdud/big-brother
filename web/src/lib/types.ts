export type Role = "admin" | "member" | "viewer";

export type ServiceType = "http" | "tcp" | "dns";

export type State = "up" | "down" | "pending" | "paused";

export interface User {
  id: string;
  email: string;
  name: string;
  role: Role;
  disabled: boolean;
  created_at: string;
  last_login_at: string;
}

export interface Project {
  id: string;
  name: string;
  slug: string;
  default_channel_id: string;
  created_at: string;
}

export interface Service {
  id: string;
  project_id: string;
  name: string;
  slug: string;
  type: ServiceType;
  url: string;
  hostname: string;
  port: number;
  interval_seconds: number;
  timeout_seconds: number;
  failure_threshold: number;
  enabled: boolean;
  channel_id: string;
  template_down: string;
  template_recovered: string;
  tags: string[];
  created_at: string;
  updated_at: string;
}

export interface Heartbeat {
  status: string;
  latency_ms: number;
  checked_at: string;
}

export interface Monitor {
  id: string;
  name: string;
  slug: string;
  type: ServiceType;
  url: string;
  hostname: string;
  port: number;
  enabled: boolean;
  tags: string[];
  interval_seconds: number;
  state: State;
  uptime_24h: number;
  uptime_30d: number;
  heartbeats: Heartbeat[];
}

export interface Status {
  state: State;
  last_change_at: string;
  last_check_at: string;
  consecutive_failures: number;
  consecutive_successes: number;
}

export interface Check {
  id: string;
  status: string;
  status_code: number;
  latency_ms: number;
  error: string;
  checked_at: string;
}

export interface Uptime {
  window: string;
  up_checks: number;
  total_checks: number;
  percent: number;
}

export interface AlertTemplate {
  trigger: "down" | "recovered";
  body: string;
  custom: boolean;
}

export interface Delivery {
  id: string;
  service_id: string;
  trigger: string;
  reason: string;
  created_at: string;
}

export interface SlackStatus {
  connected: boolean;
  team_name: string;
}

export interface ChannelOption {
  id: string;
  name: string;
}

export interface ChannelView {
  channel_id: string;
  slack_channel_id: string;
  name: string;
}

export interface ServiceInput {
  name: string;
  type: ServiceType;
  url: string;
  hostname: string;
  port: number;
  interval_seconds: number;
  timeout_seconds: number;
  failure_threshold: number;
  template_down: string;
  template_recovered: string;
  tags: string[];
}
