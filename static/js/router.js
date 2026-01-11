import { loadPage } from "./loadPage.js";
import { isAnonymous } from "./utils.js";

export function navigate(path) {
  if (isAnonymous()) {
    sessionStorage.setItem("postLoginPath", path);
    appReset("route-auth-required");
    return;
  }

  if (window.location.pathname + window.location.search === path) return;

  history.pushState({}, "", path);
  loadPage(path);
}

// Intercept internal links
export function initLinkInterceptor() {
  document.addEventListener("click", (e) => {
    const a = e.target.closest("a");
    if (!a) return;

    const href = a.getAttribute("href");
    if (!href) return;

    // ignore new tab / downloads / external links
    if (a.target === "_blank" || a.hasAttribute("download")) return;
    if (href.startsWith("http") || href.startsWith("mailto:") || href.startsWith("#")) return;
    
    // internal links only
    if (href.startsWith("/")) {
      const skipSpa = href === "/api/logout"; // usually none; keep only true server navigations
      if (skipSpa) return;

      e.preventDefault();
      navigate(href);
    }
  });
}