import { BodyText, Button, Container, Logo, TitleText } from "@/components/common";

export default function SignInPage() {
  return (
    <main className="relative flex min-h-dvh items-center bg-donna-void">
      <div
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgb(201_168_124_/_0.22),transparent_55%)]"
        aria-hidden
      />
      <Container width="narrow" className="relative z-10 py-24">
        <Logo size="lg" className="mb-10 block" />
        <TitleText className="mb-4">Sign in is next.</TitleText>
        <BodyText className="mb-8">
          Google OAuth lands in milestone M2. The landing page is ready — auth wiring comes
          right after the API skeleton.
        </BodyText>
        <Button href="/" variant="outline">
          Back to Donna
        </Button>
      </Container>
    </main>
  );
}
