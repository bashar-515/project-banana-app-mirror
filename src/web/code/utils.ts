export function makeWebSocketUrl(roomId: string, peerId: string): string {
  const url = new URL("wss://api.project-banana.com/ws")

  url.searchParams.append("roomId", roomId);
  url.searchParams.append("peerId", peerId);

  return url.toString();
}
