import { createRootRoute, createRoute, createRouter } from "@tanstack/react-router";
import { AppShell } from "./components/AppShell";
import { Login } from "./features/auth/Login";
import { Services } from "./features/services/Services";
import { Projects } from "./features/projects/Projects";
import { ServiceForm } from "./features/services/ServiceForm";
import { ServiceDetail } from "./features/services/ServiceDetail";
import { ProjectSettings } from "./features/projects/ProjectSettings";
import { AlertSettings } from "./features/alerts/AlertSettings";
import { Integrations } from "./features/integrations/Integrations";
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

const projectsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/",
  component: Projects,
});

const servicesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/services",
  component: Services,
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

const integrationsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/integrations",
  component: Integrations,
});

const usersRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/users",
  component: UsersSettings,
});

const routeTree = rootRoute.addChildren([
  loginRoute,
  appRoute.addChildren([
    projectsRoute,
    servicesRoute,
    serviceNewRoute,
    serviceEditRoute,
    serviceDetailRoute,
    settingsRoute,
    alertsRoute,
    deliveriesRoute,
    integrationsRoute,
    usersRoute,
  ]),
]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
