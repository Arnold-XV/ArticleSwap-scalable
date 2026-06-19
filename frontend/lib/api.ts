import {
  getMockArticle,
  getMockInbox,
  getMockUsers,
  submitMockArticle
} from "@/lib/mock-data";
import type {
  ArticleDetail,
  ArticleSummary,
  PipelineEvent,
  PipelineStatus,
  SubmitArticleInput,
  SubmitArticleResult,
  User
} from "@/types/articles";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export async function fetchUsers(): Promise<{ data: User[]; source: "api" | "mock" }> {
  return withMock(async () => (await request<ApiUser[]>("/users")).map(toUser), getMockUsers);
}

export async function submitArticle(
  input: SubmitArticleInput
): Promise<{ data: SubmitArticleResult; source: "api" | "mock" }> {
  return withMock(
    async () => {
      const article = await request<ApiArticle>("/articles", {
        method: "POST",
        body: JSON.stringify({
          sender_id: input.senderId,
          receiver_id: input.receiverId,
          title: input.title,
          content: input.content,
          idempotency_key: input.idempotencyKey
        })
      });
      return {
        articleId: article.id,
        status: article.status
      };
    },
    () => submitMockArticle(input)
  );
}

export async function fetchInbox(
  userId: string
): Promise<{ data: ArticleSummary[]; source: "api" | "mock" }> {
  return withMock(
    async () => (await request<ApiInboxArticle[]>(`/users/${userId}/inbox`)).map(toSummary),
    () => getMockInbox(userId)
  );
}

export async function fetchArticle(
  articleId: string
): Promise<{ data: ArticleDetail; source: "api" | "mock" }> {
  return withMock(
    async () => toArticleDetail(await request<ApiArticle>(`/articles/${articleId}`)),
    () => getMockArticle(articleId)
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

type ApiUser = {
  id: string;
  name: string;
  email: string;
};

type ApiInboxArticle = {
  id: string;
  sender_id?: string;
  senderId?: string;
  sender_name?: string;
  senderName?: string;
  title: string;
  status: string;
  created_at?: string;
  createdAt?: string;
  updated_at?: string;
  updatedAt?: string;
};

type ApiPipelineEvent = {
  service_name?: string;
  serviceName?: string;
  event_type?: string;
  eventType?: string;
  message?: string;
  created_at?: string;
  createdAt?: string;
};

type ApiProcessingResult = {
  stemmed_content?: string | null;
  stemmedContent?: string | null;
  word_frequencies?: Record<string, number> | string | null;
  wordFrequencies?: Record<string, number> | string | null;
  stemming_status?: string;
  stemmingStatus?: string;
  wordcloud_status?: string;
  wordcloudStatus?: string;
};

type ApiArticle = {
  id: string;
  sender_id?: string;
  senderId?: string;
  sender_name?: string;
  senderName?: string;
  receiver_id?: string;
  receiverId?: string;
  receiver_name?: string;
  receiverName?: string;
  title: string;
  content: string;
  content_hash?: string;
  contentHash?: string;
  status: string;
  stemmed_content?: string | null;
  stemmedContent?: string | null;
  word_frequencies?: Record<string, number> | string | null;
  wordFrequencies?: Record<string, number> | string | null;
  stemming_status?: string;
  stemmingStatus?: string;
  wordcloud_status?: string;
  wordcloudStatus?: string;
  processing?: ApiProcessingResult | null;
  events?: ApiPipelineEvent[];
  created_at?: string;
  createdAt?: string;
  updated_at?: string;
  updatedAt?: string;
};

function toUser(user: ApiUser): User {
  return {
    id: user.id,
    name: user.name,
    email: user.email
  };
}

function toSummary(article: ApiInboxArticle): ArticleSummary {
  const createdAt = article.created_at ?? article.createdAt ?? new Date().toISOString();
  return {
    id: article.id,
    senderId: article.sender_id ?? article.senderId ?? "",
    senderName: article.sender_name ?? article.senderName ?? "Pengirim",
    title: article.title,
    status: toPipelineStatus(article.status),
    createdAt,
    updatedAt: article.updated_at ?? article.updatedAt ?? createdAt
  };
}

function toArticleDetail(article: ApiArticle): ArticleDetail {
  const processing = article.processing ?? {};
  const createdAt = article.created_at ?? article.createdAt ?? new Date().toISOString();
  const stemmedContent =
    article.stemmed_content ??
    article.stemmedContent ??
    processing.stemmed_content ??
    processing.stemmedContent ??
    null;
  const wordFrequencies =
    article.word_frequencies ??
    article.wordFrequencies ??
    processing.word_frequencies ??
    processing.wordFrequencies ??
    {};

  return {
    id: article.id,
    senderId: article.sender_id ?? article.senderId ?? "",
    senderName: article.sender_name ?? article.senderName ?? "Pengirim",
    receiverId: article.receiver_id ?? article.receiverId ?? "",
    receiverName: article.receiver_name ?? article.receiverName ?? "Penerima",
    title: article.title,
    content: article.content,
    contentHash: article.content_hash ?? article.contentHash ?? "",
    status: toPipelineStatus(article.status),
    stemmedContent,
    wordFrequencies: parseWordFrequencies(wordFrequencies),
    stemmingStatus: toPipelineStatus(
      article.stemming_status ??
        article.stemmingStatus ??
        processing.stemming_status ??
        processing.stemmingStatus
    ),
    wordcloudStatus: toPipelineStatus(
      article.wordcloud_status ??
        article.wordcloudStatus ??
        processing.wordcloud_status ??
        processing.wordcloudStatus
    ),
    events: (article.events ?? []).map(toEvent),
    createdAt,
    updatedAt: article.updated_at ?? article.updatedAt ?? createdAt
  };
}

function toEvent(event: ApiPipelineEvent): PipelineEvent {
  return {
    serviceName: event.service_name ?? event.serviceName ?? "pipeline",
    eventType: event.event_type ?? event.eventType ?? "event",
    message: event.message ?? "Event pipeline diterima.",
    createdAt: event.created_at ?? event.createdAt ?? new Date().toISOString()
  };
}

function parseWordFrequencies(value: Record<string, number> | string | null | undefined) {
  if (!value) return {};
  if (typeof value !== "string") return value;

  try {
    const parsed = JSON.parse(value) as Record<string, number>;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    return {};
  }
}

function toPipelineStatus(status?: string): PipelineStatus {
  switch (status) {
    case "queued":
    case "processing":
    case "processed":
    case "degraded":
    case "failed":
      return status;
    case "done":
      return "processed";
    default:
      return "queued";
  }
}
