import {
    escapeHtml,
    formatGoDate,
    renderBadges,
    renderPostActions
  } from "./renderUtils.js";
  

/* =========================
   PUBLIC ENTRY POINT
   ========================= */

export function renderPostFromJSON(data) {
  const p = data.post;
  if (!p) return `<p>Post not found.</p>`;

  return `
    <article class="post-card" data-post-id="${p.ID}">
      ${renderPostHeader(p)}
      ${renderPostMeta(p)}
      <div class="badges">
        ${renderBadges(p.Categories)}
      </div>
      ${renderPostActions({
        id: p.ID,
        likes: p.Likes ?? 0,
        dislikes: p.Dislikes ?? 0,
        comments: p.NumComments ?? 0
      })}
      ${renderPostContent(p)}
      ${renderCommentsSection(p)}
    </article>
  `;
}

/* =========================
   PRIVATE HELPERS
   ========================= */

function renderPostHeader(p) {
  return `
    <div class="post-header">
      <h2>${escapeHtml(p.Title || "")}</h2>
    </div>
  `;
}

function renderPostMeta(p) {
  return `
    <p class="meta">
      by ${escapeHtml(p.Author || "")}
      on ${escapeHtml(formatGoDate(p.CreatedAt))}
    </p>
  `;
}

function renderPostContent(p) {
  return `
    <div class="snippet" style="white-space:pre-wrap;">
      ${escapeHtml(p.Content || "")}
    </div>
  `;
}

function renderCommentsSection(p) {
  const comments = Array.isArray(p.Comments) ? p.Comments : [];

  const commentsHtml = comments.length
    ? comments.map(renderComment).join("")
    : `<p>No comments yet.</p>`;

  return `
    <div class="comments-section show" id="comments-${p.ID}">
      <div class="comments-list" id="comments-list-${p.ID}">
        ${commentsHtml}
      </div>

      <form class="comment-form" data-post-id="${p.ID}">
        <textarea placeholder="Write a comment." required></textarea>
        <button type="submit">Post</button>
      </form>
    </div>
  `;
}

function renderComment(c) {
  return `
    <div class="comment">
      <p>
        <strong>${escapeHtml(c.Author)}</strong>:
        ${escapeHtml(c.Content)}
      </p>
      <p class="meta">${escapeHtml(formatGoDate(c.CreatedAt))}</p>

      <div class="actions">
        <button data-post-id="${c.ID}"
                data-target-type="comment"
                class="like-btn"
                data-clicked="false">
          <span class="count">${c.Likes ?? 0}</span>
        </button>

        <button data-post-id="${c.ID}"
                data-target-type="comment"
                class="dislike-btn"
                data-clicked="false">
          <span class="count">${c.Dislikes ?? 0}</span>
        </button>
      </div>
    </div>
  `;
}
