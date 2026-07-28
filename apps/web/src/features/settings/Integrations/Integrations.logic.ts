import { formatDistanceToNow } from "date-fns";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";

import type { CalendarSource } from "@/features/calendar/Calendar.types";

import {
  createICSIntegration,
  deleteICSIntegration,
  disconnectConnectedAccount,
  listCalendarSources,
  listConnectedAccounts,
  listICSIntegrations,
  startGoogleIntegration,
  startMicrosoftIntegration,
  syncICSIntegration,
  updateICSIntegration,
} from "./Integrations.api";
import type {
  ConnectedAccount,
  ICSIntegration,
  ICSIntegrationRow,
  IntegrationAccountRow,
  IntegrationProvider,
} from "./Integrations.types";

export const integrationsQueryKey = ["integrations"] as const;
export const integrationSourcesQueryKey = ["integrations", "sources"] as const;
export const icsIntegrationsQueryKey = ["integrations", "ics"] as const;

function accountTitle(account: ConnectedAccount): string {
  const name = account.display_name?.trim();
  const email = account.email?.trim();
  if (name && (!email || name !== email)) {
    return name;
  }
  return email || name || account.provider_account_id || "Connected account";
}

function accountEmailLabel(account: ConnectedAccount): string | null {
  const email = account.email?.trim();
  if (!email) {
    return null;
  }
  const name = account.display_name?.trim();
  if (name && name === email) {
    return null;
  }
  return email;
}

function accountInitials(title: string): string {
  const parts = title.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "D";
  }
  if (parts.length === 1) {
    return parts[0]!.slice(0, 2).toUpperCase();
  }
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

function syncStatusLabel(status: string): string {
  switch (status) {
    case "running":
      return "Syncing";
    case "success":
    case "succeeded":
    case "idle":
      return "Ready";
    case "partial":
      return "Partial";
    case "failed":
      return "Needs attention";
    default:
      return status || "Ready";
  }
}

function lastSyncedLabel(iso?: string | null): string {
  if (!iso) {
    return "Not synced yet";
  }
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return "Not synced yet";
  }
  return `Synced ${formatDistanceToNow(date, { addSuffix: true })}`;
}

function defaultAccountIds(sources: CalendarSource[]): Set<string> {
  const ids = new Set<string>();
  for (const source of sources) {
    if (source.is_primary_on_provider) {
      ids.add(source.connected_account_id);
    }
  }
  return ids;
}

export function toIntegrationRows(
  accounts: ConnectedAccount[],
  sources: CalendarSource[],
): IntegrationAccountRow[] {
  const defaults = defaultAccountIds(sources);
  return accounts.map((account) => {
    const title = accountTitle(account);
    return {
      ...account,
      title,
      emailLabel: accountEmailLabel(account),
      initials: accountInitials(title),
      hasDefaultCalendar: defaults.has(account.id),
      lastSyncedLabel: lastSyncedLabel(account.last_synced_at),
      syncStatusLabel: syncStatusLabel(account.calendar_sync_status),
    };
  });
}

export function toICSRows(feeds: ICSIntegration[]): ICSIntegrationRow[] {
  return feeds.map((feed) => ({
    ...feed,
    lastSyncedLabel: lastSyncedLabel(feed.last_synced_at),
    syncStatusLabel: syncStatusLabel(feed.calendar_sync_status),
    eventCountLabel:
      feed.event_count === 1 ? "1 event" : `${feed.event_count} events`,
  }));
}

export function accountsForProvider(
  rows: IntegrationAccountRow[],
  provider: IntegrationProvider,
): IntegrationAccountRow[] {
  return rows.filter((row) => row.provider === provider);
}

export function useIntegrationsController() {
  const queryClient = useQueryClient();
  const [icsName, setICSName] = useState("");
  const [icsURL, setICSURL] = useState("");
  const [icsSyncEnabled, setICSSyncEnabled] = useState(true);
  const [icsFormError, setICSFormError] = useState<string | null>(null);

  const accountsQuery = useQuery({
    queryKey: integrationsQueryKey,
    queryFn: listConnectedAccounts,
  });

  const sourcesQuery = useQuery({
    queryKey: integrationSourcesQueryKey,
    queryFn: ({ signal }) => listCalendarSources(signal),
  });

  const icsQuery = useQuery({
    queryKey: icsIntegrationsQueryKey,
    queryFn: listICSIntegrations,
  });

  useEffect(() => {
    void queryClient.invalidateQueries({ queryKey: ["calendar"] });
  }, [queryClient]);

  const invalidateAll = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: integrationsQueryKey }),
      queryClient.invalidateQueries({ queryKey: integrationSourcesQueryKey }),
      queryClient.invalidateQueries({ queryKey: icsIntegrationsQueryKey }),
      queryClient.invalidateQueries({ queryKey: ["calendar"] }),
    ]);
  };

  const disconnectMutation = useMutation({
    mutationFn: disconnectConnectedAccount,
    onSuccess: () => {
      void invalidateAll();
    },
  });

  const createICSMutation = useMutation({
    mutationFn: createICSIntegration,
    onSuccess: async () => {
      setICSName("");
      setICSURL("");
      setICSSyncEnabled(true);
      setICSFormError(null);
      await invalidateAll();
    },
    onError: (error) => {
      setICSFormError(
        error instanceof Error ? error.message : "Couldn’t add calendar URL",
      );
    },
  });

  const updateICSMutation = useMutation({
    mutationFn: ({
      id,
      sync_enabled,
    }: {
      id: string;
      sync_enabled: boolean;
    }) => updateICSIntegration(id, { sync_enabled }),
    onSuccess: () => {
      void invalidateAll();
    },
  });

  const syncICSMutation = useMutation({
    mutationFn: syncICSIntegration,
    onSuccess: () => {
      void invalidateAll();
    },
  });

  const deleteICSMutation = useMutation({
    mutationFn: deleteICSIntegration,
    onSuccess: () => {
      void invalidateAll();
    },
  });

  const rows = toIntegrationRows(
    accountsQuery.data ?? [],
    sourcesQuery.data?.sources ?? [],
  );

  return {
    rows,
    googleAccounts: accountsForProvider(rows, "google"),
    microsoftAccounts: accountsForProvider(rows, "microsoft"),
    icsFeeds: toICSRows(icsQuery.data ?? []),
    isLoading:
      accountsQuery.isLoading || sourcesQuery.isLoading || icsQuery.isLoading,
    isError: accountsQuery.isError || sourcesQuery.isError || icsQuery.isError,
    errorMessage:
      accountsQuery.error instanceof Error
        ? accountsQuery.error.message
        : sourcesQuery.error instanceof Error
          ? sourcesQuery.error.message
          : icsQuery.error instanceof Error
            ? icsQuery.error.message
            : "Couldn’t load integrations",
    refetch: () => {
      void accountsQuery.refetch();
      void sourcesQuery.refetch();
      void icsQuery.refetch();
    },
    disconnectingId: disconnectMutation.isPending
      ? (disconnectMutation.variables ?? null)
      : null,
    disconnect: (id: string) => disconnectMutation.mutateAsync(id),
    connectGoogle: startGoogleIntegration,
    connectMicrosoft: startMicrosoftIntegration,
    icsName,
    setICSName,
    icsURL,
    setICSURL,
    icsSyncEnabled,
    setICSSyncEnabled,
    icsFormError,
    icsSubmitting: createICSMutation.isPending,
    addICSFeed: () => {
      const name = icsName.trim();
      const url = icsURL.trim();
      if (!name || !url) {
        setICSFormError("Name and calendar URL are required.");
        return;
      }
      setICSFormError(null);
      createICSMutation.mutate({
        name,
        ics_url: url,
        sync_enabled: icsSyncEnabled,
      });
    },
    togglingICSId: updateICSMutation.isPending
      ? (updateICSMutation.variables?.id ?? null)
      : null,
    toggleICSSync: (id: string, sync_enabled: boolean) =>
      updateICSMutation.mutateAsync({ id, sync_enabled }),
    syncingICSId: syncICSMutation.isPending
      ? (syncICSMutation.variables ?? null)
      : null,
    syncICS: (id: string) => syncICSMutation.mutateAsync(id),
    deletingICSId: deleteICSMutation.isPending
      ? (deleteICSMutation.variables ?? null)
      : null,
    deleteICS: (id: string) => deleteICSMutation.mutateAsync(id),
  };
}
