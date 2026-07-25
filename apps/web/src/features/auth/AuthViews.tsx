"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { Button, Logo, Pill } from "@/components/common";

import { authStyles as styles } from "./Auth.styles";
import { useAuth } from "./AuthProvider";
import {
  listEnabledSignInProviders,
  SignInProvidersList,
  startSignInProvider,
  type SignInProviderId,
} from "./SignInModal";
import { startGoogleOAuth } from "./Auth.logic";

export function AuthCallback() {
  const router = useRouter();
  const params = useSearchParams();
  const { completeOAuthCallback } = useAuth();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function finish() {
      try {
        await completeOAuthCallback(params);
        if (!cancelled) {
          router.replace("/dashboard");
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Sign-in failed");
        }
      }
    }

    void finish();
    return () => {
      cancelled = true;
    };
  }, [completeOAuthCallback, params, router]);

  return (
    <div className={styles.page}>
      <div className={styles.glow} aria-hidden />
      <div className={styles.card}>
        <Logo size="sm" />
        <Pill className={styles.eyebrow} withDot>
          Signing you in
        </Pill>
        <h1 className={styles.title}>Almost there.</h1>
        {error ? (
          <>
            <p className={styles.error}>{error}</p>
            <div className={styles.actions}>
              <Button href="/sign-in" size="lg">
                Try again
              </Button>
            </div>
          </>
        ) : (
          <p className={styles.status}>Confirming your Donna session…</p>
        )}
      </div>
    </div>
  );
}

export function SignInView() {
  const { status } = useAuth();
  const router = useRouter();
  const providers = listEnabledSignInProviders();

  useEffect(() => {
    if (status === "authenticated") {
      router.replace("/dashboard");
    }
  }, [router, status]);

  const handleSelectProvider = (id: SignInProviderId) => {
    startSignInProvider(id, {
      google: startGoogleOAuth,
    });
  };

  if (status === "authenticated" || status === "loading") {
    return (
      <div className={styles.page}>
        <div className={styles.glow} aria-hidden />
        <div className={styles.card}>
          <Logo size="sm" />
          <Pill className={styles.eyebrow} withDot>
            Checking session
          </Pill>
          <h1 className={styles.title}>One moment.</h1>
          <p className={styles.status}>Looking for your Donna session…</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.glow} aria-hidden />
      <div className={styles.card}>
        <Logo size="sm" />
        <Pill className={styles.eyebrow} withDot>
          Welcome back
        </Pill>
        <h1 className={styles.title}>Sign in to Donna</h1>
        <p className={styles.body}>
          Choose how you&apos;d like to continue. Donna keeps the session
          secure — no provider tokens land in the browser.
        </p>
        <div className={styles.actions}>
          <SignInProvidersList
            providers={providers}
            onSelect={handleSelectProvider}
          />
          <Button href="/" size="lg" variant="outline">
            Back to home
          </Button>
        </div>
        <p className={styles.note}>Takes under a minute.</p>
      </div>
    </div>
  );
}
