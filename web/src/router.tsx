import { createRootRoute, createRoute, createRouter } from "@tanstack/react-router";
import { AppShell } from "./components/AppShell";
import { Login } from "./features/auth/Login";
import { Dashboard } from "./features/dashboard/Dashboard";
import { ServiceForm } from "./features/services/ServiceForm";
import { ServiceDetail } from "./features/services/ServiceDetail";
import { ProjectSettings } from "./features/projects/ProjectSettings";
import { AlertSettings } from "./features/alerts/AlertSettings";
import { SlackSettings } from "./features/slack/SlackSettings";
import { UsersSettings } from "./features/users/UsersSettings";
import { Deliveries } from "./features/deliveries/Deliveries";

const rootRoute = createRootRoute();

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: Login,
});

const appRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: "app",
  component: AppShell,
});

const dashboardRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/",
  component: Dashboard,
});

const serviceNewRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/services/new",
  component: ServiceForm,
});

const serviceEditRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/services/$service/edit",
  component: ServiceForm,
});

const serviceDetailRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/services/$service",
  component: ServiceDetail,
});

const settingsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/settings",
  component: ProjectSettings,
});

const alertsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/alerts",
  component: AlertSettings,
});

const deliveriesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/deliveries",
  component: Deliveries,
});

const slackRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/slack",
  component: SlackSettings,
});

const usersRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/users",
  component: UsersSettings,
});

const routeTree = rootRoute.addChildren([
  loginRoute,
  appRoute.addChildren([
    dashboardRoute,
    serviceNewRoute,
    serviceEditRoute,
    serviceDetailRoute,
    settingsRoute,
    alertsRoute,
    deliveriesRoute,
    slackRoute,
    usersRoute,
  ]),
]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
