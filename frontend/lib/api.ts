import {
  getMockArticle,
  getMockInbox,
  getMockUsers,
  submitMockArticle
} from "@/lib/mock-data";
import type {
  ArticleDetail,
  ArticleSummary,
  SubmitArticleInput,
  SubmitArticleResult,
  User
} from "@/types/articles";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export async function fetchUsers(): Promise<{ data: User[]; source: "api" | "mock" }> {
  return withMock(() => request<User[]>("/users"), getMockUsers);
}

export async function submitArticle(
  input: SubmitArticleInput
): Promise<{ data: SubmitArticleResult; source: "api" | "mock" }> {
  return withMock(
    () =>
      request<SubmitArticleResult>("/articles", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    () => submitMockArticle(input)
  );
}

export async function fetchInbox(
  userId: string
): Promise<{ data: ArticleSummary[]; source: "api" | "mock" }> {
  return withMock(() => request<ArticleSummary[]>(`/users/${userId}/inbox`), () =>
    getMockInbox(userId)
  );
}

export async function fetchArticle(
  articleId: string
): Promise<{ data: ArticleDetail; source: "api" | "mock" }> {
  return withMock(() => request<ArticleDetail>(`/articles/${articleId}`), () =>
    getMockArticle(articleId)
  );
}

async function withMock<T>(
  apiCall: () => Promise<T>,
  mockCall: () => T
): Promise<{ data: T; source: "api" | "mock" }> {
  try {
    return {
      data: await apiCall(),
      source: "api"
    };
  } catch {
    return {
      data: mockCall(),
      source: "mock"
    };
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    }
  });

  if (!response.ok) {
    throw new Error(`API gagal: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

