import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { usersApi } from "@/lib/api";
import { errorMessage } from "@/lib/format";
import { useSession } from "@/lib/session";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Role, User } from "@/lib/types";

const roles: Role[] = ["admin", "member", "viewer"];

export function UsersSettings() {
  const { isAdmin } = useSession();
  const queryClient = useQueryClient();
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [role, setRole] = useState<Role>("viewer");
  const [pendingRemove, setPendingRemove] = useState<User | null>(null);

  const users = useQuery({ queryKey: ["users"], queryFn: usersApi.list, enabled: isAdmin });
  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["users"] });

  const add = useMutation({
    mutationFn: () => usersApi.create(email, name, role),
    onSuccess: () => {
      toast.success("User added");
      setEmail("");
      setName("");
      setRole("viewer");
      void invalidate();
    },
    onError: (err) => toast.error(errorMessage(err)),
  });

  const update = useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: { role?: string; disabled?: boolean } }) =>
      usersApi.update(id, patch),
    onSuccess: () => void invalidate(),
    onError: (err) => toast.error(errorMessage(err)),
  });

  const remove = useMutation({
    mutationFn: (id: string) => usersApi.remove(id),
    onSuccess: () => {
      toast.success("User removed");
      setPendingRemove(null);
      void invalidate();
    },
    onError: (err) => toast.error(errorMessage(err)),
  });

  if (!isAdmin) return <p className="text-sm text-muted-foreground">Only admins can manage users.</p>;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Add user</CardTitle>
          <CardDescription>Members can be added before their first sign-in.</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            className="grid items-end gap-3 md:grid-cols-4"
            onSubmit={(event) => {
              event.preventDefault();
              if (email) add.mutate();
            }}
          >
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="user-name">Name</Label>
              <Input id="user-name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Role</Label>
              <Select value={role} onValueChange={(value) => setRole(value as Role)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {roles.map((r) => (
                    <SelectItem key={r} value={r}>
                      {r}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button type="submit" disabled={add.isPending || !email}>
              Add user
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
        </CardHeader>
        <CardContent className="px-0">
          {users.isLoading ? (
            <div className="space-y-2 px-6">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Email</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Active</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(users.data ?? []).map((user) => (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{user.email}</TableCell>
                    <TableCell>
                      <Select
                        value={user.role}
                        onValueChange={(value) => update.mutate({ id: user.id, patch: { role: value } })}
                      >
                        <SelectTrigger className="w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {roles.map((r) => (
                            <SelectItem key={r} value={r}>
                              {r}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell>
                      <Switch
                        checked={!user.disabled}
                        onCheckedChange={(checked) =>
                          update.mutate({ id: user.id, patch: { disabled: !checked } })
                        }
                      />
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        aria-label={`Remove ${user.email}`}
                        onClick={() => setPendingRemove(user)}
                      >
                        <Trash2 />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={pendingRemove !== null}
        onOpenChange={(open) => {
          if (!open) setPendingRemove(null);
        }}
        title={`Remove ${pendingRemove?.email ?? "user"}?`}
        description="They will lose access immediately."
        confirmLabel="Remove"
        pending={remove.isPending}
        onConfirm={() => {
          if (pendingRemove) remove.mutate(pendingRemove.id);
        }}
      />
    </div>
  );
}
