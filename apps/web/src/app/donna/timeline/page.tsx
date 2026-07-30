import { redirect } from "next/navigation";

/** Push / legacy Timeline deep links land on Calendar. */
export default async function DonnaTimelineDeepLinkPage({
  searchParams,
}: {
  searchParams: Promise<{ occurrence?: string }>;
}) {
  const params = await searchParams;
  const occurrence = params.occurrence?.trim();
  if (occurrence) {
    redirect(
      `/dashboard/calendar?event=${encodeURIComponent(occurrence)}`,
    );
  }
  redirect("/dashboard/calendar");
}
