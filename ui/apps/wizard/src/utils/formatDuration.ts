export function formatDuration(start?: Date | null, end?: Date | null): string {
  if (!start) return "—";
  const elapsed = (end ?? new Date()).getTime() - start.getTime();
  const seconds = Math.floor(elapsed / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}
