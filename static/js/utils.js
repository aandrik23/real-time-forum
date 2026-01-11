// A static list of 20 seed strings you want to offer
export const AVATAR_SEEDS = [
    "demo","alice","bob","carol","dave",
    "eve","frank","grace","heidi","ivan",
    "judy","mallory","nia","oscar","peggy",
    "quincy","rick","sybil","trent","victor",
];

export function fillCsrfInputs() {
    const csrfToken = getCookie("csrf_token");
    if (!csrfToken) return;
    document.querySelectorAll(".csrf_token_input").forEach((el) => {
      el.value = csrfToken;
    });
  }

export function getCookie(name) {
    const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
    return match ? decodeURIComponent(match[1]) : null;
}

export function openModal(modalEl) {
    modalEl?.classList.add('open');
}

export function closeModal(modalEl) {
    modalEl?.classList.remove('open');
}

export function isAnonymous() {
    return document.body.dataset.showLogin === "1";
  }

export function syncTheme() {
    const saved = localStorage.getItem("theme");
    const isDark = saved === "dark";
    document.documentElement.classList.toggle("dark-mode", isDark);
  
    const toggle = document.getElementById("theme-toggle");
    if (toggle) toggle.checked = isDark;
  }
  
  