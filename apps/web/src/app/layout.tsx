import type { Metadata, Viewport } from "next";
import { Instrument_Serif, Manrope } from "next/font/google";

import { ThemeProvider } from "@/components/theme";
import { AuthProvider } from "@/features/auth";
import { siteMetadata } from "@/features/landing/Landing.logic";
import { ClearDevServiceWorkers, PwaInstallPrompt } from "@/features/pwa";
import { QueryProvider } from "@/providers/QueryProvider";

import "./globals.css";

const instrumentSerif = Instrument_Serif({
  subsets: ["latin"],
  weight: "400",
  style: ["normal", "italic"],
  variable: "--font-instrument-serif",
  display: "swap",
});

const manrope = Manrope({
  subsets: ["latin"],
  variable: "--font-manrope",
  display: "swap",
});

export const metadata: Metadata = {
  ...siteMetadata,
  applicationName: "Donna",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
    title: "Donna",
  },
  icons: {
    icon: [
      { url: "/icons/donna-icon.svg", type: "image/svg+xml" },
      { url: "/icons/icon-192.png", sizes: "192x192", type: "image/png" },
      { url: "/icons/icon-512.png", sizes: "512x512", type: "image/png" },
    ],
    apple: [{ url: "/icons/apple-touch-icon.png", sizes: "180x180" }],
  },
  formatDetection: {
    telephone: false,
  },
};

export const viewport: Viewport = {
  themeColor: "#0b0b0c",
  colorScheme: "dark",
};

const themeInitScript = `
(() => {
  try {
    const key = "donna-theme";
    const stored = localStorage.getItem(key);
    const theme = stored === "light" || stored === "dark"
      ? stored
      : (window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark");
    document.documentElement.classList.remove("light", "dark");
    document.documentElement.classList.add(theme);
    document.documentElement.style.colorScheme = theme;
  } catch {}
})();
`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${instrumentSerif.variable} ${manrope.variable} dark`}
      suppressHydrationWarning
    >
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body>
        <ThemeProvider>
          <QueryProvider>
            <AuthProvider>
              <ClearDevServiceWorkers />
              {children}
              <PwaInstallPrompt />
            </AuthProvider>
          </QueryProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
