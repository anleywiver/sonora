import { useAuthStore } from "@/store/auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

export class ApiError extends Error {
  status: number;
  code: string;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

// apiFetch always sends the current access token and the refresh cookie
// (credentials: "include"); it never retries on 401 itself — callers that
// need silent-refresh-then-retry (i.e. everything behind the app shell)
// go through bootstrapSession() first (see providers.tsx).
export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const accessToken = useAuthStore.getState().accessToken;

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...init.headers,
    },
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const body = await res.json().catch(() => null);

  if (!res.ok) {
    const err = body?.error;
    throw new ApiError(res.status, err?.code ?? "unknown", err?.message ?? "Request failed");
  }

  return body?.data as T;
}
