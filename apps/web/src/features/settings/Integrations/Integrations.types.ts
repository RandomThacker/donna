export type IntegrationProvider = "google" | "microsoft" | "ics";

export type ConnectedAccount = {
  id: string;
  public_id: string;
  provider: IntegrationProvider | string;
  provider_account_id: string;
  display_name?: string | null;
  email?: string | null;
  avatar_url?: string | null;
  status: string;
  scopes: string[];
  token_expires_at?: string | null;
  last_synced_at?: string | null;
  calendar_sync_status: string;
  created_at: string;
  updated_at: string;
};

export type ICSIntegration = {
  id: string;
  public_id: string;
  provider: "ics" | string;
  name: string;
  status: string;
  sync_enabled: boolean;
  last_synced_at?: string | null;
  calendar_sync_status: string;
  event_count: number;
  created_at: string;
  updated_at: string;
};

export type IntegrationAccountRow = ConnectedAccount & {
  title: string;
  emailLabel: string | null;
  initials: string;
  hasDefaultCalendar: boolean;
  lastSyncedLabel: string;
  syncStatusLabel: string;
};

export type ICSIntegrationRow = ICSIntegration & {
  lastSyncedLabel: string;
  syncStatusLabel: string;
  eventCountLabel: string;
};
