import { useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Activity } from "lucide-react";
import { useSession } from "@/lib/session";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function Login() {
  const { state } = useSession();
  const navigate = useNavigate();

  useEffect(() => {
    if (state === "authenticated" || state === "disabled") void navigate({ to: "/" });
  }, [state, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-h1">
            <Activity className="size-5 text-primary" />
            big-brother
          </CardTitle>
          <CardDescription>Sign in to manage monitoring and alerts.</CardDescription>
        </CardHeader>
        <CardContent>
          {state === "loading" ? (
            <Skeleton className="h-8 w-full" />
          ) : (
            <Button asChild className="w-full">
              <a href="/auth/login">Sign in with Google</a>
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
