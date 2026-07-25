"use client";

import { useRouter } from "next/navigation";

import { Button, Icon } from "@/components/common";
import type { ButtonSize, ButtonVariant } from "@/components/common";

import { useAuth } from "./AuthProvider";

type AuthEntryCtaProps = {
  label: string;
  size?: ButtonSize;
  variant?: ButtonVariant;
  className?: string;
  withGoogleIcon?: boolean;
  withArrowIcon?: boolean;
};

/**
 * Landing/auth entry control. Copy stays fixed; click routes by session:
 * authenticated → dashboard, otherwise → sign-in modal.
 */
export function AuthEntryCta({
  label,
  size = "md",
  variant = "primary",
  className,
  withGoogleIcon = false,
  withArrowIcon = false,
}: AuthEntryCtaProps) {
  const { status, openSignIn } = useAuth();
  const router = useRouter();

  return (
    <Button
      size={size}
      variant={variant}
      className={className}
      iconLeft={
        withGoogleIcon ? <Icon name="google" className="h-5 w-5" /> : undefined
      }
      iconRight={
        withArrowIcon ? <Icon name="arrow" className="h-4 w-4" /> : undefined
      }
      onClick={() => {
        if (status === "authenticated") {
          router.push("/dashboard");
          return;
        }
        openSignIn();
      }}
    >
      {label}
    </Button>
  );
}
