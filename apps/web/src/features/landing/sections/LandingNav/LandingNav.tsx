"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { Container, Logo } from "@/components/common";
import { AuthEntryCta } from "@/features/auth";
import { cn } from "@/lib/cn";

import { landingNavStyles as styles } from "./LandingNav.styles";
import type { LandingNavProps } from "./LandingNav.types";

function linkIsActive(pathname: string, href: string): boolean {
  if (href.startsWith("/#") || href.startsWith("#")) {
    return pathname === "/";
  }
  if (href === "/") {
    return pathname === "/";
  }
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function LandingNav({ navLinks, getStarted }: LandingNavProps) {
  const pathname = usePathname();

  return (
    <header className={styles.header}>
      <Container>
        <div className={styles.bar}>
          <Logo size="sm" />
          <nav className={styles.links} aria-label="Primary">
            {navLinks.map((link) => {
              const active = linkIsActive(pathname, link.href);
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={cn(styles.link, active && styles.linkActive)}
                  aria-current={active ? "page" : undefined}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>
          <div className={styles.actions}>
            <AuthEntryCta label={getStarted.label} />
          </div>
        </div>
      </Container>
    </header>
  );
}
