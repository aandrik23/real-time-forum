import { closeModal, openModal, isAnonymous } from "./utils.js";
import { navigate } from "./router.js";

// DOM Elements
const sidebar = document.querySelector(".sidebar");
const toggleBtn = document.getElementById("sidebar-toggle");
const contentWrapper = document.querySelector(".container.content");
const profileBtn = document.querySelector(".sidebar-btn.profile-btn");

// ------------------ FILTER MODAL (updated for SPA) ------------------
export function initFilterModal() {
    const filterModal = document.getElementById('filterModal');
    const filterBtn = document.querySelector('.filter-btn');
    const filterForm = document.getElementById('filterForm');
  
    if (!filterBtn || !filterModal || !filterForm) return;
  
    filterBtn.addEventListener('click', () => {
      console.log("Filter button clicked!");
      openModal(filterModal);
    });
  
    filterModal.addEventListener('click', e => {
      if (e.target === filterModal) closeModal(filterModal);
    });
  
    filterModal.querySelectorAll('.modal-close').forEach(btn => {
      btn.addEventListener('click', () => closeModal(filterModal));
    });
  
    filterForm.addEventListener('submit', e => {
      e.preventDefault();
  
      const sortValue = filterForm.sort ? filterForm.sort.value : "";
      const selectedCategories = [...filterForm.querySelectorAll('input[name="category"]:checked')]
        .map(cb => cb.value);
  
      const url = new URL(window.location.href);
  
      // Keep your existing names
      url.searchParams.set('sort', sortValue);
      url.searchParams.set('categories', selectedCategories.join(','));
  
      // Also set backend-compatible params used in HomeHandler:
      if (sortValue) {
        url.searchParams.set('filter', sortValue); // HomeHandler expects "filter"
      }
      if (selectedCategories.length === 1) {
        url.searchParams.set('category', selectedCategories[0]); // HomeHandler expects single "category"
      } else {
        url.searchParams.delete('category');
      }
  
      const path = url.pathname + url.search;
  
      closeModal(filterModal);
      navigate(path);
    });
  }


// ------------------ THEME TOGGLE ------------------
export function initThemeToggle() {
    const themeToggle = document.getElementById("theme-toggle");
    if (!themeToggle) return;
  
    if (localStorage.getItem('theme') === 'dark') {
      document.documentElement.classList.add('dark-mode');
      themeToggle.checked = true;
    }
  
    themeToggle.addEventListener('change', () => {
      const isDark = themeToggle.checked;
      document.documentElement.classList.toggle('dark-mode', isDark);
      localStorage.setItem('theme', isDark ? 'dark' : 'light');
    });
  }
  
  // ------------------ SIDEBAR TOGGLE ------------------
export function initSidebarToggle() {
  if (isAnonymous()) {
    document.querySelector(".sidebar")?.remove();
    document.getElementById("sidebar-toggle")?.remove();
    return;
  }
    const collapsed = localStorage.getItem('sidebarCollapsed') === 'true';
    sidebar?.classList.toggle('collapsed', collapsed);
    contentWrapper?.classList.toggle('collapsed', collapsed);
  
    toggleBtn?.addEventListener('click', () => {
      const isCollapsed = sidebar.classList.toggle('collapsed');
      contentWrapper.classList.toggle('collapsed', isCollapsed);
      localStorage.setItem('sidebarCollapsed', isCollapsed);
    });
  }
  
  // ------------------ NAVIGATION (updated for SPA) ------------------
  export function initNavLinks() {
    const newPostBtn = document.querySelector(".newpost-btn");
    newPostBtn?.addEventListener("click", (e) => {
      e.preventDefault();
      navigate("/posts/new");
    });
  }
