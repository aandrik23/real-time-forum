import {
    escapeHtml,
    formatGoDate,
    renderBadges,
    renderPostActions
  } from "./renderUtils.js";
  

/* =========================
   PUBLIC ENTRY POINT
   ========================= */

export function renderHomeFromJSON(data) {
  return `
    ${renderWelcome(data)}
    ${renderFilterBar(data)}
    ${renderPostsSection(data)}
    ${renderDeleteModal()}
  `;
}

/* =========================
   PRIVATE HELPERS
   ========================= */

function renderWelcome(data) {
  const isUser = !!data.user;
  const username = data.username || "";

  return isUser
    ? `<p class="welcome-msg">Welcome, ${escapeHtml(username)}!</p>`
    : `<p class="welcome-msg">Welcome!</p>`;
}

function renderFilterBar(data) {
  const isUser = !!data.user;
  const cats = Array.isArray(data.categories) ? data.categories : [];

  const myButtons = isUser
    ? `
      <a class="btn" href="/home?filter=created">My Posts</a>
      <a class="btn" href="/home?filter=liked">Liked Posts</a>
    `
    : "";

  const newPostBtn = isUser
    ? `<a class="btn new-post" href="/posts/new">+ New Post</a>`
    : "";

  const categoryOptions = `
    <option value="">All Categories</option>
    ${cats.map(c => `<option value="${c.ID}">${escapeHtml(c.Name)}</option>`).join("")}
  `;

  return `
    <section class="filter-bar">
      <a class="btn" href="/home">All Posts</a>
      ${myButtons}
      <select id="categorySelect">
        ${categoryOptions}
      </select>
      ${newPostBtn}
    </section>
  `;
}

function renderPostsSection(data) {
  const isUser = !!data.user;
  const username = data.username || "";
  const posts = Array.isArray(data.posts) ? data.posts : [];

  const postsHtml = posts.length
    ? posts.map(p => renderPostCard(p, isUser, username)).join("")
    : `
      <p class="no-posts-msg">
        No posts yet.
        ${isUser ? `Why not <a href="/posts/new">create one</a>?` : ``}
      </p>
    `;

  return `
    <section class="posts-list">
      ${postsHtml}
    </section>
  `;
}

function renderPostCard(p, isUser, username) {
  const canDelete = isUser && p.Author === username;
  const comments = Array.isArray(p.Comments) ? p.Comments : [];

  const commentsHtml = comments.length
    ? comments.map(renderComment).join("")
    : `<p>No comments yet. Be the first to comment!</p>`;

  return `
    <article class="post-card" data-post-id="${p.ID}">
      <div class="post-header">
        <h2>${escapeHtml(p.Title)}</h2>

        ${canDelete ? renderDeleteButton(p.ID) : ""}
      </div>

      <p class="meta">
        by ${escapeHtml(p.Author)} on ${escapeHtml(formatGoDate(p.CreatedAt))}
      </p>

      <div class="badges">
        ${renderBadges(p.Categories)}
      </div>

      ${renderPostActions({
          id: p.ID,
          likes: p.Likes ?? 0,
          dislikes: p.Dislikes ?? 0,
          comments: p.NumComments ?? 0
      })}


      <div class="comments-section hidden" id="comments-${p.ID}">
        <div class="comments-list" id="comments-list-${p.ID}">
          ${commentsHtml}
        </div>

        <form class="comment-form" data-post-id="${p.ID}">
          <textarea placeholder="Write a comment." required></textarea>
          <button type="submit">Post</button>
        </form>
      </div>
    </article>
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

function renderDeleteButton(postId) {
  return `
    <button type="button"
            class="delete-btn"
            data-delete-post-id="${postId}"
            title="Delete Post">
      <img src="/static/img/delete.png"
           alt="Delete"
           class="delete-icon">
    </button>
  `;
}

function renderDeleteModal() {
  return `
    <div id="postDeleteModal" class="post-delete-overlay hidden">
      <div class="post-delete-modal">
        <p>Are you sure you want to delete this post?</p>
        <div class="post-delete-buttons">
          <button id="cancelPostDelete" class="btn cancel">Cancel</button>
          <button id="confirmPostDelete" class="btn delete">Yes, delete</button>
        </div>
      </div>
    </div>
  `;
}
