import { redirect } from "next/navigation";

/** Temporary stand-in until Google OAuth (M2). */
export default function SignInPage() {
  redirect("/dashboard");
}
