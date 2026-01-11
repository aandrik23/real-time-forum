import { escapeHtml } from "./renderUtils.js";

/* =========================
   PUBLIC ENTRY POINT
   ========================= */

export function renderCreatePostFromJSON(data) {
  const categories = Array.isArray(data.categories) ? data.categories : [];

  return `
    <section class="post-form-section container">
      <h1>Create New Post</h1>
      ${renderCreatePostForm(categories)}
    </section>
  `;
}

/* =========================
   PRIVATE HELPERS
   ========================= */

function renderCreatePostForm(categories) {
  return `
    <form id="createPostForm" class="post-form">
      ${renderTitleField()}
      ${renderCategorySelector(categories)}
      ${renderContentField()}
      ${renderFormButtons()}
    </form>
  `;
}

function renderTitleField() {
  return `
    <label for="title">Title</label>
    <input
      type="text"
      id="title"
      name="title"
      required
      maxlength="100"
      placeholder="Enter post title"
    >
  `;
}

function renderCategorySelector(categories) {
  return `
    <label>Categories</label>
    <div class="category-chip-group">
      ${categories.map(renderCategoryChip).join("")}
    </div>
  `;
}

function renderCategoryChip(c) {
  return `
    <label class="category-chip">
      <input type="checkbox" name="categories" value="${c.ID}">
      <span>${escapeHtml(c.Name)}</span>
    </label>
  `;
}

function renderContentField() {
  return `
    <label for="content">Content</label>
    <textarea
      id="content"
      name="content"
      rows="8"
      placeholder="Write your post here..."
      required
    ></textarea>
  `;
}

function renderFormButtons() {
  return `
    <div class="form-buttons">
      <button
        type="button"
        class="btn cancel-btn"
        id="cancelCreatePostBtn"
      >
        Cancel
      </button>

      <button
        type="submit"
        class="btn primary-btn"
      >
        Submit
      </button>
    </div>
  `;
}
