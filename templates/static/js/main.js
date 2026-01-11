import {initLoginForm, initRegisterForm, initAuthModals, initLogout} from "./auth.js";
import { initPageContent, loadPage } from "./loadPage.js";
import { initProfileDislikesRedirect, initProfileLikesRedirect, initProfilePostsRedirect } from "./profile.js";
import { initNavLinks, initSidebarToggle, initThemeToggle } from "./navigation.js";
import { initLinkInterceptor } from "./router.js";
import { fillCsrfInputs, openModal } from "./utils.js";
import { initCommentForm, initLikeButtons } from "./comments.js";
import { setAuthState } from "./auth.js";
import { onAppReset } from "./appReset.js";

onAppReset(() => {
  setAuthState(false);

  const root = document.getElementById("app-root");
  if (root) root.innerHTML = "";

  history.replaceState({}, "", "/");
  openModal(document.getElementById("loginModal"));
});
// ------------------ INITIALIZATION ------------------
document.addEventListener('DOMContentLoaded', () => {
  const loggedIn = document.body.dataset.showLogin !== "1";
  setAuthState(loggedIn);

  initAuthModals();
  initLoginForm();
  initRegisterForm();
  initLogout();
  fillCsrfInputs();
  initThemeToggle();
  if (!loggedIn) {
    document.getElementById("app-root").innerHTML = "";
    document.getElementById("loginModal")?.classList.add("open");
    return; // stop initialization
  }
  const savedSeed = localStorage.getItem('avatarSeed');
  if (savedSeed) {
    const img = document.querySelector('.profile-avatar img');
    if (img) {
      img.src = `https://api.dicebear.com/7.x/bottts/svg?seed=${encodeURIComponent(savedSeed)}`;
    }
  }
  
  document.addEventListener("click", (e) => {
    const btn = e.target.closest(".comment-toggle-btn");
    if (!btn) return;
    const postId = btn.dataset.postId;
    const section = document.getElementById(`comments-${postId}`);
    if (!section) return;
    section.classList.toggle("hidden");
    section.classList.toggle("show");
  });
  
  initLikeButtons();
  initCommentForm();
  initSidebarToggle();
  initNavLinks();
  initProfilePostsRedirect();
  initProfileLikesRedirect();
  initProfileDislikesRedirect();
  initLinkInterceptor();
  initPageContent(); // run once for initial content
  const initialPath = window.location.pathname + window.location.search;
  loadPage(initialPath);
  
});



//     // ——— new category–injection code ———
//     const rawCats = link.dataset.categories || '';
//     const cats    = rawCats ? rawCats.split(',') : [];
//     const catContainer = modal.querySelector('.modal-categories');

// // clear old ones
//     catContainer.innerHTML = '';

// // append only this post’s badges
//     cats.forEach(name => {
//         const span = document.createElement('span');
//         span.className   = 'badge';
//         span.textContent = name;
//         catContainer.appendChild(span);
//     });
// // ————————————————————————————————
