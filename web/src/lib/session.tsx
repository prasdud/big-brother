import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import { ApiError, sessionApi } from "./api";
import type { Role, User } from "./types";

type SessionState = "loading" | "authenticated" | "anonymous" | "disabled";

interface SessionValue {
  state: SessionState;
  user: User | null;
  role: Role;
  isAdmin: boolean;
  canWrite: boolean;
  signOut: () => Promise<void>;
}

const SessionContext = createContext<SessionValue | null>(null);

export function useSession(): SessionValue {
  const value = useContext(SessionContext);
  if (!value) throw new Error("useSession used outside SessionProvider");
  return value;
}

export function SessionProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<SessionState>("loading");
  const [user, setUser] = useState<User | null>(null);

  const load = useCallback(async () => {
    try {
      const me = await sessionApi.me();
      setUser(me);
      setState("authenticated");
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) {
        // Auth is not configured (local development); treat as admin.
        setUser(null);
        setState("disabled");
      } else {
        setUser(null);
        setState("anonymous");
      }
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const role: Role = state === "disabled" ? "admin" : (user?.role ?? "viewer");

  const signOut = useCallback(async () => {
    if (state === "disabled") return;
    try {
      await sessionApi.logout();
    } catch {
      // ignore; fall through to local state
    }
    setUser(null);
    setState("anonymous");
  }, [state]);

  return (
    <SessionContext.Provider
      value={{
        state,
        user,
        role,
        isAdmin: role === "admin",
        canWrite: role === "admin" || role === "member",
        signOut,
      }}
    >
      {children}
    </SessionContext.Provider>
  );
}
