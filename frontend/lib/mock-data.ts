import type {
  ArticleDetail,
  ArticleSummary,
  PipelineEvent,
  PipelineStatus,
  SubmitArticleInput,
  SubmitArticleResult,
  User
} from "@/types/articles";

export const mockUsers: User[] = [
  {
    id: "maulana-faris",
    name: "Maulana Faris Al Ghifari",
    email: "maulana.faris@example.local"
  },
  {
    id: "raditya-nathaniel",
    name: "Raditya Nathaniel Nugroho",
    email: "raditya.nathaniel@example.local"
  },
  {
    id: "ajie-armansyah",
    name: "Ajie Armansyah Sunaryo",
    email: "ajie.armansyah@example.local"
  },
  {
    id: "arnoldus-dharma",
    name: "Arnoldus Dharma Wasesa M.",
    email: "arnoldus.dharma@example.local"
  },
  {
    id: "aliya-khairun",
    name: "Aliya Khairun Nisa",
    email: "aliya.khairun@example.local"
  }
];

const now = new Date().toISOString();

const mockEvents: PipelineEvent[] = [
  {
    serviceName: "api-gateway",
    eventType: "queued",
    message: "Artikel diterima dan dimasukkan ke antrean RabbitMQ.",
    createdAt: now
  },
  {
    serviceName: "stemmer-worker",
    eventType: "processed",
    message: "Stemming selesai memakai mock fallback frontend.",
    createdAt: now
  }
];

const seedArticle: ArticleDetail = {
  id: "mock-article-1",
  senderId: "raditya-nathaniel",
  senderName: "Raditya Nathaniel Nugroho",
  receiverId: "maulana-faris",
  receiverName: "Maulana Faris Al Ghifari",
  title: "Arsitektur ArticleSwap",
  content:
    "ArticleSwap memisahkan API Gateway, RabbitMQ, worker stemming, worker word cloud, Redis, dan PostgreSQL agar sistem lebih scalable dan resilien.",
  contentHash: "mock-content-hash",
  status: "processed",
  stemmedContent:
    "articleswap pisah api gateway rabbitmq worker stem worker word cloud redis postgresql agar scalable resilien",
  wordFrequencies: {
    worker: 2,
    articleswap: 1,
    rabbitmq: 1,
    redis: 1,
    postgresql: 1,
    scalable: 1
  },
  stemmingStatus: "processed",
  wordcloudStatus: "processed",
  events: mockEvents,
  createdAt: now,
  updatedAt: now
};

let mockArticles: ArticleDetail[] = [seedArticle];

export function getMockUsers(): User[] {
  return mockUsers;
}

export function getMockInbox(userId: string): ArticleSummary[] {
  refreshMockPipeline();
  return mockArticles
    .filter((article) => article.receiverId === userId)
    .sort((left, right) => Date.parse(right.createdAt) - Date.parse(left.createdAt))
    .map(toSummary);
}

export function getMockArticle(articleId: string): ArticleDetail {
  refreshMockPipeline();
  const article = mockArticles.find((item) => item.id === articleId);
  if (!article) {
    throw new Error("Artikel mock tidak ditemukan");
  }
  return article;
}

export function submitMockArticle(input: SubmitArticleInput): SubmitArticleResult {
  const sender = mockUsers.find((user) => user.id === input.senderId);
  const receiver = mockUsers.find((user) => user.id === input.receiverId);

  if (!sender || !receiver) {
    throw new Error("Pengirim atau penerima tidak ditemukan");
  }

  const createdAt = new Date().toISOString();
  const articleId = `mock-${Date.now()}`;
  const status: PipelineStatus = "processing";
  const article: ArticleDetail = {
    id: articleId,
    senderId: sender.id,
    senderName: sender.name,
    receiverId: receiver.id,
    receiverName: receiver.name,
    title: input.title,
    content: input.content,
    contentHash: `mock-hash-${articleId}`,
    status,
    stemmedContent: null,
    wordFrequencies: {},
    stemmingStatus: "queued",
    wordcloudStatus: "queued",
    events: [
      {
        serviceName: "frontend-mock",
        eventType: "queued",
        message: "Artikel disimpan sementara di mock fallback karena API backend belum tersedia.",
        createdAt
      }
    ],
    createdAt,
    updatedAt: createdAt
  };

  mockArticles = [article, ...mockArticles];
  return {
    articleId,
    status
  };
}

function refreshMockPipeline() {
  const currentTime = Date.now();

  mockArticles = mockArticles.map((article) => {
    if (article.status !== "processing") return article;

    const ageInSeconds = (currentTime - Date.parse(article.createdAt)) / 1000;
    if (ageInSeconds < 5) {
      return {
        ...article,
        stemmingStatus: ageInSeconds >= 2 ? "processing" : article.stemmingStatus,
        wordcloudStatus: ageInSeconds >= 3 ? "processing" : article.wordcloudStatus,
        events:
          article.events.length > 1
            ? article.events
            : [
                ...article.events,
                {
                  serviceName: "frontend-mock",
                  eventType: "processing",
                  message: "Mock pipeline sedang mensimulasikan worker stemming dan word cloud.",
                  createdAt: new Date(currentTime).toISOString()
                }
              ]
      };
    }

    const stemmedContent = createMockStem(article.content);
    const wordFrequencies = createWordFrequencies(stemmedContent);
    const updatedAt = new Date(currentTime).toISOString();

    return {
      ...article,
      status: "processed",
      stemmedContent,
      wordFrequencies,
      stemmingStatus: "processed",
      wordcloudStatus: "processed",
      updatedAt,
      events: [
        ...article.events,
        {
          serviceName: "stemmer-worker",
          eventType: "processed",
          message: "Stemming mock selesai dan hasil siap ditampilkan.",
          createdAt: updatedAt
        },
        {
          serviceName: "wordcloud-worker",
          eventType: "processed",
          message: "Frekuensi kata mock selesai dihitung untuk visual word cloud.",
          createdAt: updatedAt
        }
      ]
    };
  });
}

function createMockStem(content: string) {
  const stopWords = new Set([
    "dan",
    "yang",
    "di",
    "ke",
    "dari",
    "untuk",
    "dengan",
    "agar",
    "lebih",
    "ini",
    "itu",
    "adalah"
  ]);

  return tokenize(content)
    .filter((word) => !stopWords.has(word))
    .map((word) =>
      word
        .replace(/(kan|nya|lah|pun)$/u, "")
        .replace(/^(meng|meny|men|mem|ber|ter|per|di|ke)/u, "")
    )
    .filter(Boolean)
    .join(" ");
}

function createWordFrequencies(content: string) {
  return tokenize(content).reduce<Record<string, number>>((frequencies, word) => {
    frequencies[word] = (frequencies[word] ?? 0) + 1;
    return frequencies;
  }, {});
}

function tokenize(content: string) {
  return content
    .toLowerCase()
    .replace(/[^a-z0-9\s]/g, " ")
    .split(/\s+/)
    .filter((word) => word.length > 2);
}

function toSummary(article: ArticleDetail): ArticleSummary {
  return {
    id: article.id,
    senderId: article.senderId,
    senderName: article.senderName,
    title: article.title,
    status: article.status,
    createdAt: article.createdAt,
    updatedAt: article.updatedAt
  };
}
