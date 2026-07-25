export type AuthUser = {
  id: string;
  public_id: string;
  email: string;
  email_verified: boolean;
  display_name?: string | null;
  avatar_url?: string | null;
  timezone: string;
  locale?: string | null;
  status: string;
  last_login_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type AuthStatus = "loading" | "authenticated" | "unauthenticated";

export type AuthContextValue = {
  status: AuthStatus;
  user: AuthUser | null;
  isNewUser: boolean;
  isSignInOpen: boolean;
  openSignIn: () => void;
  closeSignIn: () => void;
  signInWithGoogle: () => void;
  signOut: () => Promise<void>;
  completeOAuthCallback: (params: URLSearchParams) => Promise<void>;
  refreshSession: () => Promise<void>;
};
