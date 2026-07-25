"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import {
  clearClientAuthFlags,
  consumeNewUserFlag,
  fetchCurrentUser,
  logoutSession,
  markNewUser,
  parseOAuthCallback,
  startGoogleOAuth,
} from "./Auth.logic";
import type { AuthContextValue, AuthStatus, AuthUser } from "./Auth.types";
import {
  SignInModal,
  startSignInProvider,
  type SignInProviderId,
} from "./SignInModal";

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isNewUser, setIsNewUser] = useState(false);
  const [isSignInOpen, setIsSignInOpen] = useState(false);

  const refreshSession = useCallback(async () => {
    try {
      const me = await fetchCurrentUser();
      setUser(me);
      setStatus("authenticated");
      setIsNewUser(consumeNewUserFlag());
    } catch {
      clearClientAuthFlags();
      setUser(null);
      setStatus("unauthenticated");
      setIsNewUser(false);
    }
  }, []);

  useEffect(() => {
    void refreshSession();
  }, [refreshSession]);

  const completeOAuthCallback = useCallback(async (params: URLSearchParams) => {
    const parsed = parseOAuthCallback(params);
    if (!parsed.ok) {
      clearClientAuthFlags();
      setUser(null);
      setStatus("unauthenticated");
      throw new Error(parsed.error || "Google sign-in did not complete");
    }

    markNewUser(parsed.isNewUser);
    setIsNewUser(parsed.isNewUser);

    const me = await fetchCurrentUser();
    setUser(me);
    setStatus("authenticated");
  }, []);

  const signOut = useCallback(async () => {
    await logoutSession();
    clearClientAuthFlags();
    setUser(null);
    setStatus("unauthenticated");
    setIsNewUser(false);
  }, []);

  const openSignIn = useCallback(() => {
    setIsSignInOpen(true);
  }, []);

  const closeSignIn = useCallback(() => {
    setIsSignInOpen(false);
  }, []);

  const handleSelectProvider = useCallback((id: SignInProviderId) => {
    startSignInProvider(id, {
      google: startGoogleOAuth,
    });
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      status,
      user,
      isNewUser,
      isSignInOpen,
      openSignIn,
      closeSignIn,
      signInWithGoogle: startGoogleOAuth,
      signOut,
      completeOAuthCallback,
      refreshSession,
    }),
    [
      status,
      user,
      isNewUser,
      isSignInOpen,
      openSignIn,
      closeSignIn,
      signOut,
      completeOAuthCallback,
      refreshSession,
    ],
  );

  return (
    <AuthContext.Provider value={value}>
      {children}
      <SignInModal
        open={isSignInOpen}
        onClose={closeSignIn}
        onSelectProvider={handleSelectProvider}
      />
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return value;
}
