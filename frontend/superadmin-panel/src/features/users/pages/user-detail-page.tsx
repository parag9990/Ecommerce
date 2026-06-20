import { ArrowLeft } from "lucide-react";
import { Link, useParams } from "react-router-dom";

import { DataState } from "../../../components/ui/data-state";
import { UserProfilePanel } from "../components/user-profile-panel";
import { UserSessionList } from "../components/user-session-list";
import { useAdminUser } from "../hooks/use-admin-user";
import { useUserSessions } from "../hooks/use-user-sessions";

function displayName(fullName?: string | null): string {
  return fullName?.trim() || "Unnamed user";
}

export function UserDetailPage() {
  const { userId = "" } = useParams();
  const userQuery = useAdminUser(userId);
  const sessionsQuery = useUserSessions(userId);

  if (!userId) {
    return (
      <DataState
        tone="danger"
        title="User profile not found"
        description="The route did not include a valid user id."
      />
    );
  }

  if (userQuery.isLoading) {
    return <DataState title="Loading profile" description="Fetching the selected user profile." />;
  }

  if (userQuery.error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load profile"
        description={userQuery.error instanceof Error ? userQuery.error.message : "The profile request failed."}
        action={
          <button
            type="button"
            onClick={() => void userQuery.refetch()}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  if (!userQuery.data) {
    return <DataState tone="danger" title="User profile not found" description="No user matched this id." />;
  }

  return (
    <section className="min-h-[calc(100vh-6.5rem)] overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <Link
          to="/admin/users"
          className="mb-3 inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-950"
        >
          <ArrowLeft size={16} aria-hidden="true" />
          Back to users
        </Link>
        <div className="min-w-0">
          <h1 className="truncate text-xl font-semibold text-slate-950">
            {displayName(userQuery.data.full_name)}
          </h1>
          <p className="mt-1 break-all text-sm text-slate-600">{userQuery.data.user_id}</p>
        </div>
      </header>

      <div className="grid gap-4 p-4 xl:grid-cols-[360px_minmax(0,1fr)]">
        <UserProfilePanel user={userQuery.data} />
        <UserSessionList
          sessions={sessionsQuery.data?.sessions ?? []}
          isLoading={sessionsQuery.isLoading}
          error={sessionsQuery.error}
          onRetry={() => void sessionsQuery.refetch()}
        />
      </div>
    </section>
  );
}
