import { getCookie, isAnonymous } from "./utils.js";
import { authFetch } from "./auth.js";
import { escapeHtml, formatGoDate } from "./render/renderUtils.js";

//NEW COMMENT
let commentsBound = false;

export function initCommentForm() {
  if (isAnonymous()) return;
  if (commentsBound) return;
  commentsBound = true;

  document.addEventListener("submit", async (e) => {
    const form = e.target.closest(".comment-form");
    if (!form) return;

    e.preventDefault();

    const postId = Number(form.dataset.postId);
    const textarea = form.querySelector("textarea");
    const content = textarea?.value.trim();
    const csrfToken = getCookie("csrf_token");
    if (!postId || !content) return;

    const list =
      form.closest(".comments-section")?.querySelector(".comments-list") ||
      document.getElementById("previewCommentsList"); // for preview modal
    if (!list) return;

    // 1) optimistic comment
    const tempId = `tmp-${Date.now()}`;
    const optimistic = document.createElement("div");
    optimistic.className = "comment";
    optimistic.dataset.tempId = tempId;
    optimistic.innerHTML = `
      <p><strong>You</strong>: ${escapeHtml(content)}</p>
      <p class="meta">sending...</p>
      <div class="actions">
        <button data-post-id="0" data-target-type="comment" class="like-btn" data-clicked="false" disabled>
          <span class="count">0</span>
        </button>
        <button data-post-id="0" data-target-type="comment" class="dislike-btn" data-clicked="false" disabled>
          <span class="count">0</span>
        </button>
      </div>
    `;
    list.appendChild(optimistic);

    // clear textbox immediately
    textarea.value = "";

    // optimistic count bump (only for home feed cards where toggle exists)
    document.querySelectorAll(
      `.comment-toggle-btn[data-post-id="${postId}"] .count`
    ).forEach(span => {
      span.textContent = (Number(span.textContent) || 0) + 1;
    });

    try {
      const res = await authFetch("/api/posts/comments", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrfToken },
        body: JSON.stringify({ post_id: postId, content }),
      });
      if (!res.ok) throw new Error(await res.text());

      const saved = await res.json(); // {id, author, content, created_at, likes, dislikes}

      // 3) upgrade optimistic -> real
      optimistic.querySelector(".meta").textContent = formatGoDate(saved.created_at);

      const likeBtn = optimistic.querySelector(".like-btn");
      const dislikeBtn = optimistic.querySelector(".dislike-btn");

      likeBtn.dataset.postId = saved.id;
      dislikeBtn.dataset.postId = saved.id;
      likeBtn.disabled = false;
      dislikeBtn.disabled = false;
    } catch (err) {
      console.error("Failed to post comment:", err);

      // rollback optimistic UI
      optimistic.remove();
    }
  });
}

export function initLoadMoreComments() {
  document.addEventListener("click", async (e) => {
    const btn = e.target.closest(".load-more-comments");
    if (!btn) return;
    const postId = Number(btn.dataset.postId);
    const offset = Number(btn.dataset.offset) || 0;
    const limit = 5;

    const listId = btn.dataset.listId;
    const list = listId
      ? document.getElementById(listId)
      : document.getElementById(`comments-list-${postId}`);
    if (!list) return;

    const res = await authFetch(`/api/posts/comments?post_id=${postId}&limit=${limit}&offset=${offset}`);
    if (!res.ok) return;

    const comments = await res.json();
    if (!comments.length) {
      btn.remove();
      return;
    }

    comments.forEach(c => {
      const div = document.createElement("div");
      div.className = "comment";
      div.innerHTML = `
        <p>
          <strong>${escapeHtml(c.Author)}</strong>:
          ${escapeHtml(c.Content)}
        </p>
        <p class="meta">${(formatGoDate(c.CreatedAt))}</p>
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
        `;
      list.appendChild(div);
    });
    btn.dataset.offset = String(offset + comments.length);

    // If fewer comments than limit were returned, remove the button
    const total = Number(btn.dataset.total || 0);
    if (comments.length < limit || (total && (offset + comments.length) >= total)) {
      btn.remove();
    }
  });
}




let reactionsBound = false;

export function initLikeButtons() {
  if (isAnonymous()) return;
  if (reactionsBound) return;
  reactionsBound = true;

  document.addEventListener("click", async (e) => {
    const button = e.target.closest(".like-btn, .dislike-btn");
    if (!button) return;

    const targetId = Number(button.dataset.postId);
    const targetType = button.dataset.targetType;
    if (!targetId || !targetType) return;

    const action = button.classList.contains("like-btn") ? "like" : "dislike";
    const csrfToken = getCookie("csrf_token");

    try {
      const res = await authFetch("/api/posts/react", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrfToken },
        body: JSON.stringify({ target_type: targetType, target_id: targetId, action }),
      });

      if (!res.ok) {
        console.error("Reaction failed:", await res.text());
        return;
      }

      const data = await res.json();

      // Update ONLY inside the same comment/post container (avoid querySelector picking the first match)
      const container = button.closest(".comment, .post-card");
      if (!container) return;

      const likeBtn = container.querySelector(`.like-btn[data-target-type="${targetType}"]`);
      const dislikeBtn = container.querySelector(`.dislike-btn[data-target-type="${targetType}"]`);

      likeBtn?.querySelector(".count") && (likeBtn.querySelector(".count").textContent = data.likes);
      dislikeBtn?.querySelector(".count") && (dislikeBtn.querySelector(".count").textContent = data.dislikes);

      if (data.user_reaction === "like") {
        likeBtn?.setAttribute("data-clicked", "true");
        dislikeBtn?.setAttribute("data-clicked", "false");
      } else if (data.user_reaction === "dislike") {
        likeBtn?.setAttribute("data-clicked", "false");
        dislikeBtn?.setAttribute("data-clicked", "true");
      } else {
        likeBtn?.setAttribute("data-clicked", "false");
        dislikeBtn?.setAttribute("data-clicked", "false");
      }
    } catch (err) {
      console.error("Network error:", err);
    }
  });
}
