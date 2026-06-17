"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  fetchArticle,
  fetchInbox,
  fetchUsers,
  submitArticle
} from "@/lib/api";
import type {
  ArticleDetail,
  ArticleSummary,
  PipelineStatus,
  User
} from "@/types/articles";

const statusLabels: Record<PipelineStatus, string> = {
  queued: "Menunggu",
  processing: "Diproses",
  processed: "Selesai",
  degraded: "Sebagian Gagal",
  failed: "Gagal"
};

const emptyForm = {
  title: "",
  content: ""
};

export default function HomePage() {
  const [users, setUsers] = useState<User[]>([]);
  const [senderId, setSenderId] = useState("");
  const [receiverId, setReceiverId] = useState("");
  const [form, setForm] = useState(emptyForm);
  const [inbox, setInbox] = useState<ArticleSummary[]>([]);
  const [selectedArticle, setSelectedArticle] = useState<ArticleDetail | null>(null);
  const [dataSource, setDataSource] = useState<"api" | "mock">("mock");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLoadingInbox, setIsLoadingInbox] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(
    null
  );

  const receiverOptions = useMemo(
    () => users.filter((user) => user.id !== senderId),
    [senderId, users]
  );

  useEffect(() => {
    let ignore = false;

    async function loadUsers() {
      const result = await fetchUsers();
      if (ignore) return;
      setUsers(result.data);
      setDataSource(result.source);
      const firstUser = result.data[0]?.id ?? "";
      const secondUser = result.data[1]?.id ?? result.data[0]?.id ?? "";
      setSenderId(firstUser);
      setReceiverId(secondUser);
    }

    loadUsers();
    return () => {
      ignore = true;
    };
  }, []);

  useEffect(() => {
    if (!receiverId) return;
    loadInbox(receiverId);
  }, [receiverId]);

  async function loadInbox(userId = receiverId) {
    if (!userId) return;
    setIsLoadingInbox(true);
    try {
      const result = await fetchInbox(userId);
      setInbox(result.data);
      setDataSource(result.source);

      if (result.data.length > 0) {
        await selectArticle(result.data[0].id);
      } else {
        setSelectedArticle(null);
      }
    } catch {
      setMessage({
        type: "error",
        text: "Inbox belum bisa dimuat. Coba lagi setelah API tersedia."
      });
    } finally {
      setIsLoadingInbox(false);
    }
  }

  async function selectArticle(articleId: string) {
    try {
      const result = await fetchArticle(articleId);
      setSelectedArticle(result.data);
      setDataSource(result.source);
    } catch {
      setMessage({
        type: "error",
        text: "Detail artikel belum bisa dimuat."
      });
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage(null);

    if (!senderId || !receiverId || !form.title.trim() || !form.content.trim()) {
      setMessage({
        type: "error",
        text: "Pengirim, penerima, judul, dan isi artikel wajib diisi."
      });
      return;
    }

    if (senderId === receiverId) {
      setMessage({
        type: "error",
        text: "Pengirim dan penerima harus berbeda."
      });
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await submitArticle({
        senderId,
        receiverId,
        title: form.title.trim(),
        content: form.content.trim(),
        idempotencyKey: crypto.randomUUID()
      });
      setDataSource(result.source);
      setForm(emptyForm);
      setMessage({
        type: "success",
        text:
          result.source === "mock"
            ? "Artikel masuk ke inbox mock. Backend belum hidup, tapi flow frontend sudah siap."
            : "Artikel berhasil dikirim ke pipeline ArticleSwap."
      });
      await loadInbox(receiverId);
      await selectArticle(result.data.articleId);
    } catch {
      setMessage({
        type: "error",
        text: "Artikel belum bisa dikirim. Periksa koneksi API atau coba lagi."
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="page">
      <header className="app-header">
        <div>
          <p className="eyebrow">ArticleSwap Scalable</p>
          <h1>Pertukaran artikel dengan pipeline asynchronous</h1>
          <p className="lead">
            Kirim artikel mentah ke penerima, lalu pantau status pemrosesan stemming dan word
            cloud. Frontend ini sudah memakai kontrak API final dan akan fallback ke mock data
            selama backend belum berjalan.
          </p>
        </div>
        <div className="source-badge">
          Sumber data
          <strong>{dataSource === "api" ? "API Backend" : "Mock Fallback"}</strong>
        </div>
      </header>

      <section className="workspace" aria-label="Workspace ArticleSwap">
        <section className="panel">
          <div className="panel-header">
            <h2 className="panel-title">Kirim Artikel</h2>
            <p className="panel-subtitle">
              Bagian Faris: form submit artikel dan integrasi API dasar.
            </p>
          </div>

          <form className="form" onSubmit={handleSubmit}>
            <div className="field">
              <label htmlFor="sender">Pengirim</label>
              <select id="sender" value={senderId} onChange={(e) => setSenderId(e.target.value)}>
                {users.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="field">
              <label htmlFor="receiver">Penerima</label>
              <select
                id="receiver"
                value={receiverId}
                onChange={(e) => setReceiverId(e.target.value)}
              >
                {receiverOptions.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="field">
              <label htmlFor="title">Judul Artikel</label>
              <input
                id="title"
                value={form.title}
                onChange={(e) => setForm((current) => ({ ...current, title: e.target.value }))}
                placeholder="Contoh: Optimasi Pipeline ArticleSwap"
              />
            </div>

            <div className="field">
              <label htmlFor="content">Isi Artikel</label>
              <textarea
                id="content"
                value={form.content}
                onChange={(e) => setForm((current) => ({ ...current, content: e.target.value }))}
                placeholder="Tulis artikel yang akan dikirim ke penerima..."
              />
            </div>

            <button className="primary-button" type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Mengirim..." : "Kirim Artikel"}
            </button>

            {message ? <div className={`notice ${message.type}`}>{message.text}</div> : null}
          </form>
        </section>

        <section className="panel">
          <div className="panel-header">
            <h2 className="panel-title">Inbox Penerima</h2>
            <p className="panel-subtitle">
              Artikel mentah langsung terlihat sebelum pipeline selesai.
            </p>
          </div>

          <div className="inbox-toolbar">
            <div className="field">
              <label htmlFor="inbox-user">Lihat inbox</label>
              <select
                id="inbox-user"
                value={receiverId}
                onChange={(e) => setReceiverId(e.target.value)}
              >
                {users.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name}
                  </option>
                ))}
              </select>
            </div>
            <button
              className="secondary-button"
              type="button"
              onClick={() => loadInbox(receiverId)}
              disabled={isLoadingInbox}
            >
              {isLoadingInbox ? "Memuat" : "Refresh"}
            </button>
          </div>

          <div className="inbox-list">
            {inbox.length === 0 ? (
              <div className="empty-state">Belum ada artikel masuk untuk user ini.</div>
            ) : (
              inbox.map((article) => (
                <button
                  className={`article-item ${
                    selectedArticle?.id === article.id ? "active" : ""
                  }`}
                  key={article.id}
                  type="button"
                  onClick={() => selectArticle(article.id)}
                >
                  <h3>{article.title}</h3>
                  <div className="article-meta">
                    <span>Dari {article.senderName}</span>
                    <StatusBadge status={article.status} />
                  </div>
                </button>
              ))
            )}
          </div>
        </section>

        <section className="panel detail-panel">
          {selectedArticle ? <ArticleDetailPanel article={selectedArticle} /> : <DetailEmpty />}
        </section>
      </section>
    </main>
  );
}

function ArticleDetailPanel({ article }: { article: ArticleDetail }) {
  const wordEntries = Object.entries(article.wordFrequencies).slice(0, 8);

  return (
    <div className="detail">
      <div className="detail-title-row">
        <div>
          <h2>{article.title}</h2>
          <p className="panel-subtitle">
            Dari {article.senderName} untuk {article.receiverName}
          </p>
        </div>
        <StatusBadge status={article.status} />
      </div>

      <div className="detail-grid">
        <div className="info-box">
          <h3>Artikel Mentah</h3>
          <p className="article-content">{article.content}</p>
        </div>

        <div className="info-box">
          <h3>Hasil Stemming</h3>
          <p>
            {article.stemmedContent ??
              "Menunggu stemmer worker. Placeholder ini siap diisi dari backend."}
          </p>
        </div>

        <div className="info-box">
          <h3>Word Cloud</h3>
          {wordEntries.length > 0 ? (
            <div className="word-placeholder">
              {wordEntries.map(([word, count]) => (
                <span className="word-chip" key={word}>
                  {word} {count}
                </span>
              ))}
            </div>
          ) : (
            <p>Menunggu wordcloud worker. Bagian visual final akan dilanjutkan Raditya.</p>
          )}
        </div>

        <div className="info-box">
          <h3>Status Pipeline</h3>
          <div className="timeline">
            {article.events.length > 0 ? (
              article.events.map((event, index) => (
                <div className="timeline-item" key={`${event.serviceName}-${index}`}>
                  <strong>
                    {event.serviceName} - {event.eventType}
                  </strong>
                  <span>{event.message}</span>
                </div>
              ))
            ) : (
              <p>Timeline pipeline belum tersedia.</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function DetailEmpty() {
  return (
    <div className="detail-empty">
      Pilih artikel dari inbox untuk melihat isi mentah, status pipeline, hasil stemming, dan
      placeholder word cloud.
    </div>
  );
}

function StatusBadge({ status }: { status: PipelineStatus }) {
  return <span className={`status ${status}`}>{statusLabels[status]}</span>;
}

