import { renderHomeFromJSON } from "./render/home.js";
import { renderCreatePostFromJSON } from "./render/createPost.js";
import { renderPostFromJSON } from "./render/post.js";
import { renderNavbarCategories } from "./render/navbar.js";
import { renderProfileFromJSON } from "./render/profile.js";

import { initDeletePostModal, initPostPreviewModal } from "./posts.js";
import { initProfileModal } from "./profile.js";
import { initFilterModal } from "./navigation.js";
import { initLikeButtons, initCommentForm } from "./comments.js";

import { getCookie } from "./utils.js";
import { navigate } from "./router.js";
import { fillCsrfInputs } from "./utils.js";
import { setAuthState, isLoggedIn, authFetch } from "./auth.js";
import { appReset } from "./appReset.js";

let isLoading = false;

window.addEventListener('popstate', () => {
  const path = window.location.pathname + window.location.search;
  loadPage(path);
});

async function loadHomeAPI(path) {
    
    const url = new URL("/api/home", window.location.origin);
  
    // Copy query params from the current SPA path to the API call
    const qIndex = path.indexOf("?");
    if (qIndex !== -1) {
      const qs = new URLSearchParams(path.slice(qIndex + 1));
      for (const [k, v] of qs.entries()) url.searchParams.set(k, v);
    }
  
    const res = await authFetch(url.toString(), {
      method: "GET",
      credentials: "include",
      headers: { Accept: "application/json" },
    });
    
    if (!res.ok) throw new Error(`Home API failed: ${res.status}`);
    
    const data = await res.json();
    setAuthState(!!data.user);
    const list = document.getElementById("categoriesList");
    if (list) {
      list.innerHTML = renderNavbarCategories(data.categories);
    }
    const appRoot = document.getElementById("app-root");
    if (!appRoot) return;
  
    appRoot.innerHTML = renderHomeFromJSON(data); // you create this renderer
    fillCsrfInputs();
    initPageContent();
  }
  
export async function loadHomeInitial() {
    // avoid history/popstate interaction entirely
    await loadHomeAPI(window.location.pathname + window.location.search);
  }
  
// ------------------ CORE SPA LOADER ------------------
export async function loadPage(path) {
    // AUTH GATE
    if (!isLoggedIn()) {
      sessionStorage.setItem("postLoginPath", path);
      appReset("route-auth-required");
      return;
    }
  
  if (isLoading) return;
  isLoading = true;

  try {

    const pathname = path.split("?")[0];

    if (pathname === "/" || pathname === "/home") {
      await loadHomeAPI(pathname === "/" ? ("/home" + (path.includes("?") ? path.slice(path.indexOf("?")) : "")) : path);
      return;
    }

    if (pathname === "/profile") {
      await loadProfileAPI();
      return;
    }

    if (pathname === "/posts/new") {
      await loadCreatePostAPI();
      return;
    }

    if (pathname.startsWith("/post/")) {
      const id = Number(pathname.split("/post/")[1]);
      await loadPostAPI(id);
      return;
    }
    
    fillCsrfInputs();
    initPageContent();
  } catch (err) {
    console.error('SPA navigation failed:', err);
  
    const appRoot = document.getElementById("app-root");
    if (appRoot) {
      appRoot.innerHTML = `<p>Something went wrong loading this page.</p>`;
    }
  } finally {
    isLoading = false;
  }
  
}


// ------------------ PER-PAGE CONTENT INIT ------------------
export function initPageContent() {
  initFilterModal();
  initProfileModal();
  initPostPreviewModal();
  initDeletePostModal();
  initLikeButtons();
  initCommentForm();
}

async function loadCreatePostAPI() {
    const res = await authFetch("/api/posts/new", {
      method: "GET",
      credentials: "include",
      headers: { "Accept": "application/json" },
    });
  
    if (!res.ok) throw new Error(`Create init API failed: ${res.status}`);
  
    const data = await res.json();
    setAuthState(true);
    const appRoot = document.getElementById("app-root");
    if (!appRoot) return;
  
    appRoot.innerHTML = renderCreatePostFromJSON(data);
    initCreatePostBindings();
    fillCsrfInputs();
    initPageContent(); // keeps likes/comments/etc safe (won't do anything on this page)
  }
  
function initCreatePostBindings() {
    const form = document.getElementById("createPostForm");
    const cancelBtn = document.getElementById("cancelCreatePostBtn");
    const sel = document.getElementById("categorySelect");
    if (sel) {
      sel.addEventListener("change", () => {
        const v = sel.value;
        navigate(v ? `/home?category=${encodeURIComponent(v)}` : "/home");
      });
    }
    if (!form) return;
  
    cancelBtn?.addEventListener("click", () => {
      const path = "/home";
      navigate(path);
    });
  
    form.addEventListener("submit", async (e) => {
      e.preventDefault();
  
      const title = form.querySelector("#title")?.value.trim();
      const content = form.querySelector("#content")?.value.trim();
      const selected = [...form.querySelectorAll('input[name="categories"]:checked')]
        .map(cb => Number(cb.value))
        .filter(n => Number.isFinite(n) && n > 0);
  
      if (!title || !content) return;
  
      try {
        const csrf = getCookie("csrf_token");
  
        const res = await authFetch("/api/posts", {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": csrf
          },
          body: JSON.stringify({
            title,
            content,
            categories: selected
          })
        });
  
        if (!res.ok) {
          console.error("Create post failed:", await res.text());
          return;
        }
  
        const out = await res.json(); // { ok, post_id }
        const path = `/post/${out.post_id}`;
  
        navigate(path);
      } catch (err) {
        console.error("Create post request failed:", err);
      }
    });
  }

async function loadPostAPI(id) {
    const res = await authFetch(`/api/post?id=${encodeURIComponent(id)}`, {
      method: "GET",
      credentials: "include",
      headers: { Accept: "application/json" },
    });
  
    if (!res.ok) throw new Error(`Post API failed: ${res.status}`);
    
    const data = await res.json();
    setAuthState(true);
  
    const appRoot = document.getElementById("app-root");
    if (!appRoot) return;
  
    appRoot.innerHTML = renderPostFromJSON(data);
    fillCsrfInputs();
    initPageContent();
  }
  
async function loadProfileAPI() {
    const res = await authFetch("/api/profile", {
      method: "GET",
      credentials: "include",
      headers: { "Accept": "application/json" },
    });
  
    if (!res.ok) throw new Error(`Profile API failed: ${res.status}`);
  
    const data = await res.json();
    setAuthState(true);
  
    const appRoot = document.getElementById("app-root");
    if (!appRoot) return;
  
    appRoot.innerHTML = renderProfileFromJSON(data);
    fillCsrfInputs();
    initPageContent();
  }
  