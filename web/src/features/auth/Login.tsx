import { useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useSession } from "../../lib/session";
import { Button, Card, CardBody, Spinner } from "../../components/ui";

export function Login() {
  const { state } = useSession();
  const navigate = useNavigate();

  useEffect(() => {
    if (state === "authenticated" || state === "disabled") void navigate({ to: "/" });
  }, [state, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50">
      <Card className="w-96">
        <CardBody className="space-y-4 text-center">
          <h1 className="text-lg font-semibold text-slate-900">big-brother</h1>
          <p className="text-sm text-slate-600">Sign in to manage monitoring and alerts.</p>
          {state === "loading" ? (
            <Spinner />
          ) : (
            <a href="/auth/login">
              <Button className="w-full">Sign in with Google</Button>
            </a>
          )}
        </CardBody>
      </Card>
    </div>
  );
}
