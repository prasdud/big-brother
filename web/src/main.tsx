import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { router } from "./router";
import { SessionProvider } from "./lib/session";
import { ProjectProvider } from "./lib/project";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import "./index.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, refetchOnWindowFocus: false },
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <SessionProvider>
          <ProjectProvider>
            <RouterProvider router={router} />
            <Toaster theme="dark" position="bottom-right" />
          </ProjectProvider>
        </SessionProvider>
      </TooltipProvider>
    </QueryClientProvider>
  </StrictMode>,
);
