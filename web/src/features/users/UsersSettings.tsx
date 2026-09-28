import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { usersApi } from "../../lib/api";
import { errorMessage } from "../../lib/format";
import { useSession } from "../../lib/session";
import { useToast } from "../../components/Toast";
import { Button, Card, CardBody, CardHeader, ErrorText, Field, Input, Select, Spinner } from "../../components/ui";
import type { Role } from "../../lib/types";

const roles: Role[] = ["admin", "member", "viewer"];

export function UsersSettings() {
  const { isAdmin } = useSession();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [role, setRole] = useState<Role>("viewer");
  const [error, setError] = useState("");

  const users = useQuery({ queryKey: ["users"], queryFn: usersApi.list, enabled: isAdmin });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["users"] });

  const add = useMutation({
    mutationFn: () => usersApi.create(email, name, role),
    onSuccess: () => {
      toast.push("success", "User added");
      setEmail("");
      setName("");
      setRole("viewer");
      void invalidate();
    },
    onError: (err) => setError(errorMessage(err)),
  });

  const update = useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: { role?: string; disabled?: boolean } }) =>
      usersApi.update(id, patch),
    onSuccess: () => void invalidate(),
    onError: (err) => toast.push("error", errorMessage(err)),
  });

  const remove = useMutation({
    mutationFn: (id: string) => usersApi.remove(id),
    onSuccess: () => {
      toast.push("success", "User removed");
      void invalidate();
    },
    onError: (err) => toast.push("error", errorMessage(err)),
  });

  if (!isAdmin) return <p className="text-sm text-slate-600">Only admins can manage users.</p>;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader title="Add user" />
        <CardBody>
          <form
            className="grid grid-cols-1 gap-3 md:grid-cols-4"
            onSubmit={(e) => {
              e.preventDefault();
              setError("");
              add.mutate();
            }}
          >
            <Field label="Email">
              <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
            </Field>
            <Field label="Name">
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </Field>
            <Field label="Role">
              <Select value={role} onChange={(e) => setRole(e.target.value as Role)}>
                {roles.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </Select>
            </Field>
            <div className="flex items-end">
              <Button type="submit" disabled={add.isPending || !email}>
                Add
              </Button>
            </div>
          </form>
          <div className="mt-2">
            <ErrorText>{error}</ErrorText>
          </div>
        </CardBody>
      </Card>

      <Card>
        <CardHeader title="Members" />
        <CardBody className="p-0">
          {users.isLoading ? (
            <div className="p-4">
              <Spinner />
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-200 text-left text-xs uppercase text-slate-500">
                  <th className="px-4 py-2">Email</th>
                  <th className="px-4 py-2">Role</th>
                  <th className="px-4 py-2">Status</th>
                  <th className="px-4 py-2 text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {(users.data ?? []).map((user) => (
                  <tr key={user.id} className="border-b border-slate-100 last:border-0">
                    <td className="px-4 py-2 text-slate-800">{user.email}</td>
                    <td className="px-4 py-2">
                      <Select
                        className="w-32"
                        value={user.role}
                        onChange={(e) => update.mutate({ id: user.id, patch: { role: e.target.value } })}
                      >
                        {roles.map((r) => (
                          <option key={r} value={r}>
                            {r}
                          </option>
                        ))}
                      </Select>
                    </td>
                    <td className="px-4 py-2 text-slate-600">{user.disabled ? "disabled" : "active"}</td>
                    <td className="px-4 py-2 text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="secondary"
                          onClick={() => update.mutate({ id: user.id, patch: { disabled: !user.disabled } })}
                        >
                          {user.disabled ? "Enable" : "Disable"}
                        </Button>
                        <Button
                          variant="danger"
                          onClick={() => {
                            if (window.confirm(`Remove ${user.email}?`)) remove.mutate(user.id);
                          }}
                        >
                          Remove
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardBody>
      </Card>
    </div>
  );
}
