"use client";

import { useState } from "react";

import { Icon, type IconName } from "@/components/common";

import { useIntegrationsController } from "./Integrations.logic";
import { integrationsStyles as styles } from "./Integrations.styles";
import type {
  ICSIntegrationRow,
  IntegrationAccountRow,
  IntegrationProvider,
} from "./Integrations.types";

function providerLabel(provider: IntegrationProvider): string {
  switch (provider) {
    case "google":
      return "Google";
    case "microsoft":
      return "Microsoft";
    case "ics":
      return "Calendar URL (ICS)";
    default:
      return provider;
  }
}

function providerIcon(provider: IntegrationProvider): IconName {
  switch (provider) {
    case "google":
      return "google";
    case "microsoft":
      return "microsoft";
    case "ics":
      return "calendar";
    default:
      return "calendar";
  }
}

function AccountAvatar({
  avatarUrl,
  initials,
}: {
  avatarUrl?: string | null;
  initials: string;
}) {
  const [failed, setFailed] = useState(false);
  const showImage = Boolean(avatarUrl) && !failed;

  return (
    <span className={styles.avatar} aria-hidden>
      {showImage ? (
        <img
          src={avatarUrl!}
          alt=""
          className={styles.avatarImage}
          referrerPolicy="no-referrer"
          onError={() => setFailed(true)}
        />
      ) : (
        initials
      )}
    </span>
  );
}

function ProviderSection({
  provider,
  accounts,
  onConnect,
  onDisconnect,
  disconnectingId,
}: {
  provider: IntegrationProvider;
  accounts: IntegrationAccountRow[];
  onConnect: () => void;
  onDisconnect: (id: string) => void;
  disconnectingId: string | null;
}) {
  return (
    <section className={styles.section} aria-labelledby={`${provider}-heading`}>
      <div className={styles.sectionHead}>
        <div>
          <h2 id={`${provider}-heading`} className={styles.sectionTitle}>
            {providerLabel(provider)}
          </h2>
          <p className={styles.sectionHint}>
            Calendar accounts stay separate from how you sign in to Donna.
          </p>
        </div>
        <button type="button" className={styles.connect} onClick={onConnect}>
          <Icon name={providerIcon(provider)} className="h-4 w-4" />
          {accounts.length === 0
            ? `Connect ${providerLabel(provider)}`
            : "Connect another account"}
        </button>
      </div>

      {accounts.length === 0 ? (
        <p className={styles.empty}>No {providerLabel(provider)} accounts yet.</p>
      ) : (
        <ul className={styles.list}>
          {accounts.map((account) => (
            <li key={account.id} className={styles.row}>
              <div className={styles.rowMain}>
                <AccountAvatar
                  avatarUrl={account.avatar_url}
                  initials={account.initials}
                />
                <div className={styles.rowMeta}>
                  <p className={styles.accountName}>{account.title}</p>
                  {account.emailLabel ? (
                    <p className={styles.accountEmail}>{account.emailLabel}</p>
                  ) : null}
                  <div className={styles.meta}>
                    <span>{providerLabel(provider)}</span>
                    <span>{account.syncStatusLabel}</span>
                    <span>{account.lastSyncedLabel}</span>
                    {account.hasDefaultCalendar ? (
                      <span className={styles.badge}>Default calendar</span>
                    ) : null}
                  </div>
                </div>
              </div>
              <button
                type="button"
                className={styles.disconnect}
                disabled={disconnectingId === account.id}
                onClick={() => {
                  void onDisconnect(account.id);
                }}
              >
                {disconnectingId === account.id ? "Disconnecting…" : "Disconnect"}
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function ICSSection({
  feeds,
  name,
  url,
  syncEnabled,
  formError,
  submitting,
  onNameChange,
  onURLChange,
  onSyncEnabledChange,
  onSubmit,
  onToggleSync,
  onSync,
  onDelete,
  togglingId,
  syncingId,
  deletingId,
}: {
  feeds: ICSIntegrationRow[];
  name: string;
  url: string;
  syncEnabled: boolean;
  formError: string | null;
  submitting: boolean;
  onNameChange: (value: string) => void;
  onURLChange: (value: string) => void;
  onSyncEnabledChange: (value: boolean) => void;
  onSubmit: () => void;
  onToggleSync: (id: string, enabled: boolean) => void;
  onSync: (id: string) => void;
  onDelete: (id: string) => void;
  togglingId: string | null;
  syncingId: string | null;
  deletingId: string | null;
}) {
  return (
    <section className={styles.section} aria-labelledby="ics-heading">
      <div className={styles.sectionHead}>
        <div>
          <div className={styles.titleWithHelp}>
            <h2 id="ics-heading" className={styles.sectionTitle}>
              Calendar URL (ICS)
            </h2>
            <span
              className={styles.help}
              title="Works with Outlook, Apple Calendar, Google published calendars, and other iCalendar-compatible services."
              aria-label="Works with Outlook, Apple Calendar, Google published calendars, and other iCalendar-compatible services."
            >
              ?
            </span>
          </div>
          <p className={styles.sectionHint}>
            Paste a public or private calendar feed URL. Donna imports events the
            same way as Google.
          </p>
        </div>
      </div>

      <form
        className={styles.form}
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit();
        }}
      >
        <div className={styles.formFields}>
          <div className={styles.formRow}>
            <label className={styles.label} htmlFor="ics-name">
              Name
            </label>
            <input
              id="ics-name"
              className={styles.input}
              value={name}
              onChange={(event) => onNameChange(event.target.value)}
              placeholder="Angel One Work Calendar"
              autoComplete="off"
            />
          </div>
          <div className={styles.formRow}>
            <label className={styles.label} htmlFor="ics-url">
              Calendar URL
            </label>
            <input
              id="ics-url"
              className={styles.input}
              value={url}
              onChange={(event) => onURLChange(event.target.value)}
              placeholder="https://… or webcal://…"
              autoComplete="off"
              spellCheck={false}
            />
          </div>
        </div>
        <label className={styles.checkboxRow}>
          <input
            type="checkbox"
            checked={syncEnabled}
            onChange={(event) => onSyncEnabledChange(event.target.checked)}
          />
          Enable sync
        </label>
        <div className={styles.formActions}>
          <button type="submit" className={styles.connect} disabled={submitting}>
            <Icon name="calendar" className="h-4 w-4" />
            {submitting ? "Adding…" : "Add calendar feed"}
          </button>
          {formError ? <p className={styles.formError}>{formError}</p> : null}
        </div>
      </form>

      {feeds.length === 0 ? (
        <p className={styles.empty}>No calendar URLs yet.</p>
      ) : (
        <ul className={styles.list}>
          {feeds.map((feed) => (
            <li key={feed.id} className={styles.row}>
              <div className={styles.rowMain}>
                <span className={styles.avatar} aria-hidden>
                  <Icon name="calendar" className="h-4 w-4" />
                </span>
                <div className={styles.rowMeta}>
                  <p className={styles.accountName}>{feed.name}</p>
                  <div className={styles.meta}>
                    <span>Calendar URL (ICS)</span>
                    <span>{feed.syncStatusLabel}</span>
                    <span>{feed.lastSyncedLabel}</span>
                    <span>{feed.eventCountLabel}</span>
                    <span>{feed.sync_enabled ? "Sync on" : "Sync off"}</span>
                  </div>
                </div>
              </div>
              <div className={styles.rowActions}>
                <button
                  type="button"
                  className={styles.secondaryAction}
                  disabled={togglingId === feed.id}
                  onClick={() => {
                    void onToggleSync(feed.id, !feed.sync_enabled);
                  }}
                >
                  {feed.sync_enabled ? "Disable sync" : "Enable sync"}
                </button>
                <button
                  type="button"
                  className={styles.secondaryAction}
                  disabled={syncingId === feed.id}
                  onClick={() => {
                    void onSync(feed.id);
                  }}
                >
                  {syncingId === feed.id ? "Syncing…" : "Sync now"}
                </button>
                <button
                  type="button"
                  className={styles.disconnect}
                  disabled={deletingId === feed.id}
                  onClick={() => {
                    void onDelete(feed.id);
                  }}
                >
                  {deletingId === feed.id ? "Removing…" : "Delete"}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

export function Integrations() {
  const ctrl = useIntegrationsController();

  return (
    <div className={styles.page}>
      <header className={styles.intro}>
        <h1 className={styles.pageTitle}>Integrations</h1>
        <p className={styles.body}>
          Connect Google or a Calendar URL (ICS). Signing in never wires these
          up automatically — you choose what Donna can see.
        </p>
      </header>

      {ctrl.isLoading ? (
        <>
          <div className={styles.skeleton} />
          <div className={styles.skeleton} />
          <div className={styles.skeleton} />
        </>
      ) : null}

      {ctrl.isError ? (
        <div className={styles.error} role="alert">
          <p className={styles.errorTitle}>Couldn’t load integrations</p>
          <p className={styles.errorBody}>{ctrl.errorMessage}</p>
          <button type="button" className={styles.retry} onClick={ctrl.refetch}>
            Try again
          </button>
        </div>
      ) : null}

      {!ctrl.isLoading && !ctrl.isError ? (
        <>
          <ProviderSection
            provider="google"
            accounts={ctrl.googleAccounts}
            onConnect={ctrl.connectGoogle}
            onDisconnect={(id) => {
              void ctrl.disconnect(id);
            }}
            disconnectingId={ctrl.disconnectingId}
          />
          {/* Microsoft integration disabled for now.
          <ProviderSection
            provider="microsoft"
            accounts={ctrl.microsoftAccounts}
            onConnect={ctrl.connectMicrosoft}
            onDisconnect={(id) => {
              void ctrl.disconnect(id);
            }}
            disconnectingId={ctrl.disconnectingId}
          />
          */}
          <ICSSection
            feeds={ctrl.icsFeeds}
            name={ctrl.icsName}
            url={ctrl.icsURL}
            syncEnabled={ctrl.icsSyncEnabled}
            formError={ctrl.icsFormError}
            submitting={ctrl.icsSubmitting}
            onNameChange={ctrl.setICSName}
            onURLChange={ctrl.setICSURL}
            onSyncEnabledChange={ctrl.setICSSyncEnabled}
            onSubmit={ctrl.addICSFeed}
            onToggleSync={(id, enabled) => {
              void ctrl.toggleICSSync(id, enabled);
            }}
            onSync={(id) => {
              void ctrl.syncICS(id);
            }}
            onDelete={(id) => {
              void ctrl.deleteICS(id);
            }}
            togglingId={ctrl.togglingICSId}
            syncingId={ctrl.syncingICSId}
            deletingId={ctrl.deletingICSId}
          />
        </>
      ) : null}
    </div>
  );
}
