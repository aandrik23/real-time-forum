import { escapeHtml } from "./renderUtils.js";

export function renderNavbarCategories(categories = []) {
  return categories.map(c => `
    <li>
      <a href="/home?category=${encodeURIComponent(c.ID)}">
        ${escapeHtml(c.Name)}
      </a>
    </li>
  `).join("");
}
