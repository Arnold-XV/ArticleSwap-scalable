export type PipelineStatus = "queued" | "processing" | "processed" | "degraded" | "failed";

export type User = {
  id: string;
  name: string;
  email: string;
};

export type ArticleSummary = {
  id: string;
  senderId: string;
  senderName: string;
  title: string;
  status: PipelineStatus;
  createdAt: string;
  updatedAt: string;
};

export type PipelineEvent = {
  serviceName: string;
  eventType: string;
  message: string;
  createdAt: string;
};

export type ArticleDetail = {
  id: string;
  senderId: string;
  senderName: string;
  receiverId: string;
  receiverName: string;
  title: string;
  content: string;
  contentHash: string;
  status: PipelineStatus;
  stemmedContent: string | null;
  wordFrequencies: Record<string, number>;
  stemmingStatus: PipelineStatus;
  wordcloudStatus: PipelineStatus;
  events: PipelineEvent[];
  createdAt: string;
  updatedAt: string;
};

export type SubmitArticleInput = {
  senderId: string;
  receiverId: string;
  title: string;
  content: string;
  idempotencyKey: string;
};

export type SubmitArticleResult = {
  articleId: string;
  status: string;
};

