import { request } from "./http";

import type {
  MockRecord,
  HealthResponse,
  MockListResponse,
  MockApiResponse,
} from "../types/mock";

export async function getHealth(signal?: AbortSignal) {
  return request<HealthResponse>("/v1/health", { signal });
}

export async function registerMock(payload: MockRecord, signal?: AbortSignal) {
  return request<MockApiResponse>("/v1/__mock", {
    method: "POST",
    body: payload,
    signal,
  });
}

export async function uploadMockConfig(file: File, signal?: AbortSignal) {
  const form = new FormData();
  form.append("config.json", file);

  return request<MockApiResponse>("/v1/__mock/upload", {
    method: "POST",
    body: form,
    signal,
  });
}

export async function listMocks(signal?: AbortSignal) {
  return request<MockListResponse>("/v1/__mock", { signal });
}

export async function clearMocks(signal?: AbortSignal) {
  return request<MockApiResponse>("/v1/__mock/all", {
    method: "DELETE",
    signal,
  });
}

export async function deleteMockByMethod(
  method: string,
  path: string,
  signal?: AbortSignal,
) {
  return request<MockApiResponse>(
    `/v1/__mock/${method}?path=${encodeURIComponent(path)}`,
    {
      method: "DELETE",
      signal,
    },
  );
}
