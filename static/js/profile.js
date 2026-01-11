import { openModal, closeModal, AVATAR_SEEDS, isAnonymous, getCookie } from "./utils.js";
import { navigate } from "./router.js";
import { authFetch } from "./auth.js";

// Profile modal
export function initProfileModal() {
  if (isAnonymous()) return;

  const editBtn           = document.querySelector('.profile-edit-btn');
  const editModal         = document.getElementById('editProfileModal');
  const avatarSelectModal = document.getElementById('avatarSelectModal');
  const avatarPreview     = document.querySelector('.avatar-preview');
  const avatarPreviewImg  = document.getElementById('avatarPreviewImg');
  const avatarSeedInput   = document.getElementById('edit-avatarSeed');

  // If we're not on the profile page, just bail out quietly
  if (!editBtn || !editModal || !avatarSelectModal || !avatarPreview || !avatarPreviewImg || !avatarSeedInput) {
    return;
  }

  const avatarGrid = avatarSelectModal.querySelector('.avatar-grid');
  if (!avatarGrid) return;

  let gridBuilt = false;

  // -------------------- OPEN EDIT PROFILE MODAL --------------------
  editBtn.addEventListener('click', () => {
    // pull from header
    const currentName = document.querySelector('.profile-username').textContent.trim();
    const currentBio  = document.querySelector('.profile-bio').textContent.trim();
    const avatarEl    = document.querySelector('.profile-avatar img');
    const avatarURL   = avatarEl.src;
    const seed        = new URL(avatarURL).searchParams.get('seed') || '';

    // set form fields
    document.getElementById('edit-username').value = currentName;
    document.getElementById('edit-bio').value      = currentBio;
    avatarSeedInput.value                          = seed;
    avatarPreviewImg.src                           = avatarURL;

    openModal(editModal);
  });

  // Close Edit Profile if you click the dark backdrop
  editModal.addEventListener('click', e => {
    if (e.target === editModal) {
      closeModal(editModal);
    }
  });

  // Allow any [data-modal-close] inside the edit modal to close it
  editModal.querySelectorAll('.modal-close')
    .forEach(btn => btn.addEventListener('click', () => closeModal(editModal)));

  // -------------------- BUILD AVATAR GRID ONCE --------------------
  function buildAvatarGrid() {
    avatarGrid.innerHTML = '';
    AVATAR_SEEDS.forEach(seed => {
      const img = document.createElement('img');
      img.src          = `https://api.dicebear.com/7.x/bottts/svg?seed=${encodeURIComponent(seed)}`;
      img.dataset.seed = seed;
      img.className    = 'avatar-thumb';
      if (seed === avatarSeedInput.value) img.classList.add('selected');
      avatarGrid.appendChild(img);
    });
    gridBuilt = true;
  }

  // Clicking preview → open picker
  avatarPreview.addEventListener('click', () => {
    if (!gridBuilt) buildAvatarGrid();
    openModal(avatarSelectModal);
  });

  // Pick an avatar
  avatarGrid.addEventListener('click', e => {
    const thumb = e.target.closest('.avatar-thumb');
    if (!thumb) return;

    // update hidden input + preview
    avatarSeedInput.value = thumb.dataset.seed;
    avatarPreviewImg.src  = thumb.src;

    // persist selection in localStorage
    localStorage.setItem('avatarSeed', thumb.dataset.seed);

    avatarGrid.querySelectorAll('.avatar-thumb.selected')
      .forEach(el => el.classList.remove('selected'));
    thumb.classList.add('selected');

    closeModal(avatarSelectModal);
  });

  // Close picker if clicking backdrop or close-btn
  avatarSelectModal.addEventListener('click', e => {
    if (e.target === avatarSelectModal) closeModal(avatarSelectModal);
  });
  avatarSelectModal.querySelectorAll('.modal-close')
    .forEach(btn => btn.addEventListener('click', () => closeModal(avatarSelectModal)));

  // -------------------- SAVE CHANGES (SPA-style) --------------------
    const form = document.getElementById('editProfileForm');
    if (!form) return;
    form.addEventListener('submit', async e => {
    e.preventDefault();
    console.log("Submitting profile form...");

    const name       = document.getElementById('edit-username').value.trim();
    const bio        = document.getElementById('edit-bio').value.trim();
    const avatarSeed = document.getElementById('edit-avatarSeed').value.trim();

    try {
      const csrfToken = getCookie('csrf_token');
      const res = await authFetch('/api/profile/update', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': csrfToken
        },
        body: JSON.stringify({ username: name, bio, avatarSeed })
      });

      const errorEl = form.querySelector(".form-error");
      if (errorEl) errorEl.textContent = "";
      
      if (!res.ok) {
        const payload = await res.json().catch(() => null);
        const msg = payload?.error || `Update failed (${res.status})`;
        if (errorEl) errorEl.textContent = msg;
        return;
      }
      if (errorEl) errorEl.textContent = "";

      //  No full reload: update the page DOM directly
      const out = await res.json().catch(() => null);
      if (!out) return;
      
      const usernameEl = document.querySelector('.profile-username');
      const bioEl      = document.querySelector('.profile-bio');
      const avatarEl   = document.querySelector('.profile-avatar img');
      
      if (usernameEl) usernameEl.textContent = ` ${out.username}`;
      if (bioEl) bioEl.textContent = out.bio || '';
      if (avatarEl) {
        avatarEl.src = `https://api.dicebear.com/7.x/bottts/svg?seed=${encodeURIComponent(out.avatarSeed || 'default')}`;
      }
      
      closeModal(editModal);

    } catch (err) {
      console.error("Network error:", err);
    }
  });
}

export function initProfileLikesRedirect() {
  const btn = document.querySelector('.user-likes-link');
  if (!btn) return;

  btn.addEventListener('click', (e) => {
    e.preventDefault();
    const path = '/home?filter=liked';

    if (isAnonymous()) {
      navigate(path)
    } else {
      navigate(path);
      history.pushState({ path }, '', path);
    }
  });
}

export function initProfileDislikesRedirect() {
  const btn = document.querySelector('.user-dislikes-link');
  if (!btn) return;

  btn.addEventListener('click', (e) => {
    e.preventDefault();
    const path = '/home?filter=disliked';
    navigate(path);

  });
}

// ------------------ POPSTATE (Back/Forward) ------------------

export function initProfilePostsRedirect() {
  const btn = document.querySelector('.user-posts-link');
  if (!btn) return;

  btn.addEventListener('click', (e) => {
    e.preventDefault();
    const path = '/home?filter=created';

    if ((isAnonymous())) {
      navigate(path);
    } else {
      navigate(path);
      history.pushState({ path }, '', path);
    }
  });
}

