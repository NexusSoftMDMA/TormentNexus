import { redirect } from "next/navigation";

export default async function ProjectIndex(props: { params: Promise<{ path: string }> }) {
  const { path } = await props.params;
  redirect(`/projects/${path}/activity`);
}
