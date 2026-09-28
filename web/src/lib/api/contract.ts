// Compile-time checks that the frontend's runtime types stay compatible with
// the published OpenAPI contract (web/src/lib/api/generated/schema.d.ts).
//
// The generated schema is produced from internal/api/openapi.json by
// `npm run gen:api`. If a field is renamed or its type changes in the spec and
// not here, these assertions fail the build.
import type { components } from "./generated/schema";
import type {
  AlertTemplate,
  ChannelOption,
  ChannelView,
  Check,
  Delivery,
  Project,
  Service,
  SlackStatus,
  Status,
  Uptime,
  User,
} from "../types";

type Schemas = components["schemas"];

type Assert<T extends true> = T;

export type ProjectContract = Assert<Project extends Schemas["Project"] ? true : false>;
export type ServiceContract = Assert<Service extends Schemas["Service"] ? true : false>;
export type StatusContract = Assert<Status extends Schemas["Status"] ? true : false>;
export type CheckContract = Assert<Check extends Schemas["Check"] ? true : false>;
export type UptimeContract = Assert<Uptime extends Schemas["Uptime"] ? true : false>;
export type AlertTemplateContract = Assert<AlertTemplate extends Schemas["AlertTemplate"] ? true : false>;
export type DeliveryContract = Assert<Delivery extends Schemas["Delivery"] ? true : false>;
export type UserContract = Assert<User extends Schemas["User"] ? true : false>;
export type SlackStatusContract = Assert<SlackStatus extends Schemas["SlackStatus"] ? true : false>;
export type ChannelOptionContract = Assert<ChannelOption extends Schemas["ChannelOption"] ? true : false>;
export type ChannelViewContract = Assert<ChannelView extends Schemas["Channel"] ? true : false>;
