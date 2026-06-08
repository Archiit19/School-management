export function parseTokenClaims(token) {
  if (!token) return {};
  try {
    const base64 = token.split(".")[1]?.replace(/-/g, "+").replace(/_/g, "/");
    if (!base64) return {};
    const json = atob(base64);
    return JSON.parse(json);
  } catch {
    return {};
  }
}
