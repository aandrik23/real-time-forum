/* =========================
   TEXT / DATE
   ========================= */

   export function escapeHtml(s) {
    return String(s)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }
  
  export function formatGoDate(dateValue) {
    const d = new Date(dateValue);
    if (Number.isNaN(d.getTime())) return "";
  
    return d.toLocaleDateString("en-GB", {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  }
  
  /* =========================
     SHARED RENDER HELPERS
     ========================= */
  
  export function renderBadges(categories = []) {
    return categories
      .map(c => `<span class="badge">${escapeHtml(c.Name)}</span>`)
      .join("");
  }
  
  export function renderPostActions({ id, likes = 0, dislikes = 0, comments = 0 }) {
    return `
      <div class="actions">
        <div class="reaction-group">
          <button
            data-post-id="${id}"
            data-target-type="post"
            class="like-btn"
            data-clicked="false"
          >
            <span class="count">${likes}</span>
          </button>
        </div>
  
        <div class="reaction-group">
          <button
            data-post-id="${id}"
            data-target-type="post"
            class="dislike-btn"
            data-clicked="false"
          >
            <span class="count">${dislikes}</span>
          </button>
        </div>
  
        <div class="reaction-group">
          <button
            data-post-id="${id}"
            class="comment-toggle-btn"
          >
            <span class="count">${comments}</span>
          </button>
        </div>
      </div>
    `;
  }
  