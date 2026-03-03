import pb from "./pocketbase";

function authHeaders(extra?: HeadersInit): HeadersInit {
  return {
    Authorization: `Bearer ${pb.authStore.token}`,
    ...extra,
  };
}

export function apiFetch(url: string, init: RequestInit = {}): Promise<Response> {
  return fetch(url, {
    ...init,
    headers: authHeaders(init.headers as HeadersInit),
  });
}

export function apiGet(url: string): Promise<Response> {
  return apiFetch(url);
}

export function apiPost(url: string, body: unknown): Promise<Response> {
  return apiFetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

export function apiPut(url: string, body: unknown): Promise<Response> {
  return apiFetch(url, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}
