import { escapeHtml, formatGoDate, renderBadges } from "./renderUtils.js";


/* =========================
   PUBLIC ENTRY POINT
   ========================= */

export function renderProfileFromJSON(data) {
    return `
      ${renderProfileHeader(data)}
      ${renderProfileActivity(data)}
      ${renderProfileModals(data)}
    `;
  }

/* =========================
   PRIVATE HELPERS
   ========================= */
function renderProfileHeader(data) {
    const username = data.username || "";
    const bio = data.bio || "";
    const avatar = data.avatar || "default";
    const stats = data.stats || {};
  
    return `
      <div class="profile-page container">
        <div class="profile-header">
          <div class="profile-avatar">
            <img
              src="https://api.dicebear.com/7.x/bottts/svg?seed=${encodeURIComponent(avatar)}"
              alt="User Avatar"
            />
          </div>
  
          <div class="profile-info">
            <h2 class="profile-username">${escapeHtml(username)}</h2>
            <p class="profile-bio">${escapeHtml(bio)}</p>
  
            <div class="profile-stats">
              <button class="user-posts-link clickable-stat" type="button">
                <strong>${stats.PostCount ?? stats.postCount ?? 0}</strong> Posts
              </button>
              <button class="user-likes-link clickable-stat" type="button">
                <strong>${stats.LikesGiven ?? stats.likesGiven ?? 0}</strong> Likes
              </button>
              <button class="user-dislikes-link clickable-stat" type="button">
                <strong>${stats.DislikesGiven ?? stats.dislikesGiven ?? 0}</strong> Dislikes
              </button>
            </div>
  
            <button class="btn pill profile-edit-btn">Edit Profile</button>
          </div>
        </div>
      </div>
    `;
  }
  
function renderProfileActivity(data) {
    const posts = Array.isArray(data.posts) ? data.posts : [];
  
    const postsHtml = posts.length
      ? posts.map(renderProfilePostItem).join("")
      : `<li><em>No recent posts yet.</em></li>`;
  
    return `
      <div class="profile-activity">
        <h3>Recent Posts</h3>
        <ul class="profile-post-list">
          ${postsHtml}
        </ul>
      </div>
    `;
  }
  
function renderProfilePostItem(p) {
    return `
      <li>
        <a href="#"
           class="profile-post-title"
           data-id="${p.ID}"
           data-title="${escapeHtml(p.Title || "")}"
           data-content="${escapeHtml(p.Content || "")}"
           data-date="${escapeHtml(formatGoDate(p.CreatedAt))}"
           data-likes="${p.Likes ?? 0}"
           data-dislikes="${p.Dislikes ?? 0}"
           data-comments="${p.NumComments ?? 0}">
          ${escapeHtml(p.Title || "")}
        </a>
        <span class="profile-post-date">
          ${escapeHtml(formatGoDate(p.CreatedAt))}
        </span>
      </li>
    `;
  }
  
function renderProfileModals(data) {
    const categories = Array.isArray(data.categories) ? data.categories : [];
    const previewCats = renderBadges(categories);
  
    return `
      ${renderEditProfileModal()}
      ${renderAvatarSelectModal()}
      ${renderPostPreviewModal(previewCats)}
    `;
  }
  
function renderEditProfileModal() {
    return `
      <div id="editProfileModal" class="modal-overlay">
        <div class="modal-card">
          <button class="modal-close" data-modal-close>&times;</button>
          <h2>Edit Profile</h2>
  
          <form id="editProfileForm" class="auth-form">
            <label for="edit-username">Username</label>
            <input type="text" id="edit-username" name="username">
  
            <label for="edit-bio">Bio</label>
            <textarea id="edit-bio" name="bio" rows="3"></textarea>
  
            <input type="hidden" id="edit-avatarSeed" name="avatarSeed">
  
            <label>Avatar</label>
            <div class="avatar-preview" style="margin:1rem 0; cursor:pointer;">
              <img id="avatarPreviewImg"
                   src=""
                   alt="Avatar preview"
                   style="width:72px;height:72px;border-radius:50%;">
              <p class="small">Click avatar to change</p>
            </div>
  
            <div class="form-error" style="color:#ff6b6b; margin:.5rem 0;"></div>
            <button type="submit" class="btn pill">Save Changes</button>
          </form>
        </div>
      </div>
    `;
  }
  
function renderAvatarSelectModal() {
    return `
      <div id="avatarSelectModal" class="modal-overlay">
        <div class="modal-card avatar-picker-card">
          <button class="modal-close" data-modal-close>&times;</button>
          <h2>Select Your Avatar</h2>
          <div class="avatar-grid"></div>
        </div>
      </div>
    `;
  }
  
function renderPostPreviewModal(previewCats) {
    return `
      <div id="postPreviewModal" class="modal-overlay">
        <div class="modal-card post-card" id="modal-post-card">
          <button class="modal-close" data-modal-close>&times;</button>
  
          <div class="post-header">
            <h2><a id="previewTitle" href="#"></a></h2>
            <p class="meta" id="previewDate"></p>
          </div>
  
          <div class="badges">
            ${previewCats}
          </div>
  
          <div class="snippet" id="previewContent" style="white-space: pre-wrap;"></div>
  
          <div class="actions">
            <div class="reaction-group">
              <button class="like-btn" data-clicked="false">
                <span class="count" id="previewLikes">0</span>
              </button>
              <button class="dislike-btn" data-clicked="false">
                <span class="count" id="previewDislikes">0</span>
              </button>
              <button class="comment-toggle-btn">
                <span class="count" id="previewComments">0</span>
              </button>
            </div>
          </div>
  
          <div class="comments-section hidden" id="previewCommentsSection">
            <div class="comments-list" id="previewCommentsList"></div>
  
            <form class="comment-form" id="previewCommentForm" data-post-id="">
              <textarea placeholder="Write a comment..." required></textarea>
              <button type="submit">Post</button>
            </form>
          </div>
        </div>
      </div>
    `;
  }