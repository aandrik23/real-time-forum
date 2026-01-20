import { authFetch } from "./auth.js";
import { onAppReset } from "./appReset.js";
import { escapeHtml } from "./render/renderUtils.js";
import { isAnonymous } from "./utils.js";
import wsManager from "./ws.js";

// -----------------------------
// App reset
// -----------------------------
onAppReset(() => {
  resetChatState();
});

export function resetChatState() {
  wsManager.resetAll();

  threads = [];
  suggestedUsers = [];
  activeChatUser = null;
  activeMessages = [];
  paging = { oldestId: null, loading: false, exhausted: false };

  pendingMessages.clear();
  deliveryStatusMap.clear();
  lastReadByUser.clear();

  document.getElementById("chat-panel")?.remove();
  document.getElementById("chat-sidebar")?.remove();
}
// -----------------------------
// DM state
// -----------------------------

let threads = [];              // left list: [{other_user_id, other_username, online, last_message_at, ...}]
let suggestedUsers = [];       // only when threads empty: [{user_id, username, online, ...}]
let activeChatUser = null;     // { id, username }
let activeMessages = [];       // current chat messages (oldest -> newest)

let paging = {
  oldestId: null,
  loading: false,
  exhausted: false
};

const CURRENT_USER_ID = Number(document.body.dataset.userId);
const pendingMessages = new Map(); // client_msg_id -> payload
const deliveryStatusMap = new Map(); // messageID -> deliveredAt
const lastReadByUser = new Map();

// -----------------------------
// Helpers
// -----------------------------
function toThreadUserListItems() {
  // If threads exist, they are the list. If none, use suggested users list.
  if (threads.length > 0) {
    return threads.map(t => ({
      id: t.other_user_id,
      username: t.other_username,
      online: !!t.online,
      lastMessageAt: t.last_message_at || 0,
      lastMessageBody: t.last_message_body || "",
      unreadCount: t.unread_count || 0
    }));
  }

  return suggestedUsers.map(u => ({
    id: u.user_id,
    username: u.username,
    online: !!u.online,
    lastMessageAt: 0,
    lastMessageBody: ""
  }));
}

function sortUsers(users) {
  return users.sort((a, b) => {
    // Online users first
    if (a.online !== b.online) return a.online ? -1 : 1;

    // Then by last message time (desc)
    if (a.lastMessageAt && b.lastMessageAt) return b.lastMessageAt - a.lastMessageAt;

    // Users with messages before those without
    if (a.lastMessageAt && !b.lastMessageAt) return -1;
    if (!a.lastMessageAt && b.lastMessageAt) return 1;

    // Finally by username
    return a.username.localeCompare(b.username);
  });
}

function fmtTime(unixSec) {
  const d = new Date(unixSec * 1000);
  // Keep it simple; you can style later
  return d.toLocaleString();
}

// basic throttle (no spam)
function throttle(fn, waitMs) {
  let last = 0;
  let timer = null;

  return (...args) => {
    const now = Date.now();
    const remaining = waitMs - (now - last);

    if (remaining <= 0) {
      last = now;
      fn(...args);
      return;
    }

    if (!timer) {
      timer = setTimeout(() => {
        timer = null;
        last = Date.now();
        fn(...args);
      }, remaining);
    }
  };
}

function isChatPanelMinimized() {
  return document.getElementById("chat-panel")?.classList.contains("minimized") || false;
}

let chatSidebarRaf = null;

function updateChatSidebarStop() {
  const sidebar = document.getElementById("chat-sidebar");
  const footer = document.querySelector(".footer");
  if (!sidebar || !footer) return;

  const navbarVar = getComputedStyle(document.documentElement).getPropertyValue("--navbar-h");
  const navbarPx = parseFloat(navbarVar) || 0;
  const sidebarHeight = sidebar.getBoundingClientRect().height || (window.innerHeight - navbarPx);

  const footerTop = footer.getBoundingClientRect().top + window.scrollY;
  const stopTop = footerTop - sidebarHeight;
  const sidebarDocTop = window.scrollY + navbarPx;

  if (sidebarDocTop >= stopTop) {
    sidebar.classList.add("chat-stop");
    sidebar.style.top = `${Math.max(stopTop, navbarPx)}px`;
  } else {
    sidebar.classList.remove("chat-stop");
    sidebar.style.top = "";
  }
}

function scheduleChatSidebarStopUpdate() {
  if (chatSidebarRaf) return;
  chatSidebarRaf = requestAnimationFrame(() => {
    chatSidebarRaf = null;
    updateChatSidebarStop();
  });
}

// -----------------------------
// DOM creation
// -----------------------------
function createChatSidebar() {
  if (isAnonymous()) return;
  const root = document.getElementById("chat-root");
  root.innerHTML = `
      <div id="chat-sidebar">
        <div id="chat-header">
          <span>Messages</span>
          <button id="chat-sidebar-toggle" aria-label="Minimize chat sidebar">—</button>
        </div>
        <div id="chat-user-list"></div>
      </div>
    `;

  document.getElementById("chat-sidebar-toggle").addEventListener("click", () => {
    document.getElementById("chat-sidebar").classList.toggle("collapsed");

    // if sidebar collapses, also move panel flush right
    const panel = document.getElementById("chat-panel");
    if (panel) panel.classList.toggle("sidebar-collapsed");

    scheduleChatSidebarStopUpdate();
  });

  scheduleChatSidebarStopUpdate();
}


function renderUsers(users) {
  const list = document.getElementById("chat-user-list");
  list.innerHTML = "";

  users.forEach(user => {
    const div = document.createElement("div");
    div.className = "chat-user";
    div.dataset.userId = String(user.id);
    div.innerHTML = `
        <div class="chat-user-left">
          <span class="chat-username">
            ${escapeHtml(user.username)}
            ${user.unreadCount > 0 ? `<span class="unread-badge">${user.unreadCount}</span>` : ``}
          </span>
          ${user.lastMessageBody ? `<div class="chat-last">${escapeHtml(user.lastMessageBody)}</div>` : ``}
        </div>
        <div class="chat-user-right">
          ${user.lastMessageAt ? `<div class="chat-last-time">${fmtTime(user.lastMessageAt)}</div>` : ``}
          <span class="chat-status ${user.online ? "chat-online" : "chat-offline"}"></span>
        </div>
      `;

    div.addEventListener("click", () => {
      document.querySelectorAll(".chat-user.active")
        .forEach(el => el.classList.remove("active"));

      div.classList.add("active");
      openChat(user);
    });
    list.appendChild(div);
  });
}

function openChat(user) {
  activeChatUser = { id: user.id, username: user.username };
  // clear unread badge locally (UI-only)
  const t = threads.find(t => t.other_user_id === user.id);
  if (t) {
    t.unread_count = 0;
  }
  renderUsers(sortUsers(toThreadUserListItems()));
  activeMessages = [];
  paging = { oldestId: null, loading: false, exhausted: false };

  let panel = document.getElementById("chat-panel");
  if (panel) panel.remove();

  panel = document.createElement("div");
  panel.id = "chat-panel";
  panel.innerHTML = `
      <div id="chat-panel-header">
        <span>Chat with ${escapeHtml(user.username)}</span>
        <div class="chat-panel-actions">
          <button id="chat-minimize" aria-label="Minimize chat">—</button>
          <button id="chat-close" aria-label="Close chat">×</button>
        </div>
      </div>
      <div id="chat-messages"></div>
      <div id="chat-input-area">
        <input id="chat-input" type="text" placeholder="Type a message..." />
        <button id="chat-send">Send</button>
      </div>
    `;

  document.body.appendChild(panel);
  // minimize chat

  document.getElementById("chat-minimize").addEventListener("click", () => {
    panel.classList.toggle("minimized");

    // If un-minimizing (no longer has 'minimized' class)
    if (!panel.classList.contains("minimized")) {
      clearUnreadForActiveChat();
      sendReadReceipt();
    }

    renderMessages();
  });

  // close chat
  document.getElementById("chat-close").addEventListener("click", () => {
    panel.remove();
    activeChatUser = null;
    activeMessages = [];
    paging = { oldestId: null, loading: false, exhausted: false };

    // clear selected highlight
    document.querySelectorAll(".chat-user.active")
      .forEach(el => el.classList.remove("active"));
  });

  const messagesEl = document.getElementById("chat-messages");
  messagesEl.innerHTML = `<p style="opacity:.6">Loading...</p>`;

  // Scroll handler (throttled)
  messagesEl.addEventListener(
    "scroll",
    throttle(() => {
      if (!activeChatUser) return;
      if (paging.loading || paging.exhausted) return;

      // When near top, load more
      if (messagesEl.scrollTop <= 20) {
        loadMoreMessages();
      }
    }, 250)
  );

  // Send handlers
  document.getElementById("chat-send").addEventListener("click", sendActiveMessage);
  document.getElementById("chat-input").addEventListener("keydown", (e) => {
    if (e.key === "Enter") sendActiveMessage();
  });

  // Load initial 10
  loadInitialMessages();
}

function renderMessages() {
  const el = document.getElementById("chat-messages");
  if (!el) return;

  if (activeMessages.length === 0) {
    el.innerHTML = `<p style="opacity:.6">No messages yet</p>`;
    return;
  }

  el.innerHTML = activeMessages
    .map(m => {
      const when = fmtTime(m.created_at);
      const body = escapeHtml(m.body);

      const isMine = m.sender_id === CURRENT_USER_ID;

      let status = "";
      const otherID = activeChatUser?.id;
      const lastRead = lastReadByUser.get(otherID) || 0;

      const isRead = isMine && m.id <= lastRead;
      const isDelivered = isMine && deliveryStatusMap.has(m.id);

      if (isRead && !isChatPanelMinimized()) {
        status = "Read";
      } else if (isDelivered) {
        status = "Delivered";
      } else {
        status = "Sent";
      }

      return `
        <div class="chat-msg ${isMine ? "mine" : "theirs"}">
          <div class="chat-msg-bubble">
            <div class="chat-msg-body">${body}</div>
            <div class="chat-msg-time">
              ${when}
              ${isMine ? `<span class="chat-msg-status">${status}</span>` : ""}
            </div>
          </div>
        </div>
      `;
    })
    .join("");
}


function appendMessage(msgObj) {
  // msgObj: {id, sender_id, sender_username, body, created_at}
  activeMessages.push(msgObj);
  renderMessages();

  // Scroll to bottom on new message (only if this chat is active)
  const el = document.getElementById("chat-messages");
  if (el) el.scrollTop = el.scrollHeight;
}

function prependMessages(msgs) {
  // msgs are oldest->newest (your backend step 10)
  const el = document.getElementById("chat-messages");
  const prevHeight = el ? el.scrollHeight : 0;

  activeMessages = msgs.concat(activeMessages);
  renderMessages();

  // Keep scroll position stable after prepending
  if (el) {
    const newHeight = el.scrollHeight;
    el.scrollTop = newHeight - prevHeight;
  }
}

// -----------------------------
// HTTP API
// -----------------------------
async function loadThreads() {
  const res = await authFetch("/api/dm/threads", { method: "GET" });
  if (!res.ok) return;

  const data = await res.json();

  threads = Array.isArray(data.threads) ? data.threads : [];
  suggestedUsers = Array.isArray(data.suggested_users) ? data.suggested_users : [];

  const listItems = sortUsers(toThreadUserListItems());
  renderUsers(listItems);
}

function sendReadReceipt() {
  if (!activeChatUser || activeMessages.length === 0) return;
  if (isChatPanelMinimized()) return;

  const lastMsg = activeMessages[activeMessages.length - 1];

  // DO NOT mark read if I sent the last message
  if (lastMsg.sender_id === CURRENT_USER_ID) return;

  wsManager.send("dm", {
    type: "dm_read",
    conversation_with: activeChatUser.id,
    last_read_msg_id: lastMsg.id
  });
}



async function loadInitialMessages() {
  paging.loading = true;
  paging.exhausted = false;
  paging.oldestId = null;

  try {
    const res = await authFetch(`/api/dm/messages?user_id=${encodeURIComponent(activeChatUser.id)}`, { method: "GET" });
    if (!res.ok) return;
    const data = await res.json();
    const msgs = Array.isArray(data.messages) ? data.messages : [];

    lastReadByUser.set(activeChatUser.id, data.last_read_msg_id || 0);

    // Merge global delivery status into each message
    msgs.forEach(msg => {
      msg.delivered = deliveryStatusMap.get(msg.id) || false;
    });

    activeMessages = msgs;
    renderMessages();

    if (msgs.length > 0) {
      paging.oldestId = msgs[0].id; // oldest
    } else {
      paging.exhausted = true;
    }

    // Scroll to bottom after initial load
    const el = document.getElementById("chat-messages");
    if (el) el.scrollTop = el.scrollHeight;

    // mark as read now that user opened the chat
    sendReadReceipt();
  } finally {
    paging.loading = false;
  }
}


async function loadMoreMessages() {
  if (!activeChatUser) return;
  if (!paging.oldestId) {
    paging.exhausted = true;
    return;
  }

  paging.loading = true;
  try {
    const url = `/api/dm/messages?user_id=${encodeURIComponent(activeChatUser.id)}&before_id=${encodeURIComponent(paging.oldestId)}`;
    const res = await authFetch(url, { method: "GET" });
    if (!res.ok) return;

    const data = await res.json();
    const msgs = Array.isArray(data.messages) ? data.messages : [];

    if (msgs.length === 0) {
      paging.exhausted = true;
      return;
    }

    // Update oldestId to the oldest message in the newly received batch
    paging.oldestId = msgs[0].id;

    // Prepend
    prependMessages(msgs);
  } finally {
    paging.loading = false;
  }
}

// -----------------------------
// WebSocket
// -----------------------------
function handleWsMessage(ev) {
  let msg;
  try {
    msg = JSON.parse(ev.data);
  } catch {
    return;
  }

  switch (msg.type) {

    case "dm_ack": {
      const { client_msg_id } = msg;
      pendingMessages.delete(client_msg_id);
      break;
    }

    case "dm_new": {
      const { conversation_with, message } = msg;

      const isActiveChat =
        activeChatUser && activeChatUser.id === conversation_with;

      if (isActiveChat) {
        appendMessage(message);
        // Only send read receipt if chat is NOT minimized
        if (!isChatPanelMinimized()) {
          sendReadReceipt();
        } else {
          // Chat is active but minimized - increment unread count
          const t = threads.find(
            t => t.other_user_id === conversation_with
          );
          if (t) {
            t.unread_count = (t.unread_count || 0) + 1;
          }
          renderUsers(sortUsers(toThreadUserListItems()));
        }
      } else {
        //increment unread count locally
        const t = threads.find(
          t => t.other_user_id === conversation_with
        );
        if (t) {
          t.unread_count = (t.unread_count || 0) + 1;
        }
        renderUsers(sortUsers(toThreadUserListItems()));
      }

      break;
    }

    case "dm_delivery": {
      const { server_msg_id, delivered } = msg;
      if (delivered) {
        deliveryStatusMap.set(server_msg_id, true);
        renderMessages();
      }
      break;
    }

    case "dm_read": {
      const { conversation_with, last_read_msg_id } = msg;
      lastReadByUser.set(conversation_with, last_read_msg_id);
      renderMessages();
      break;
    }

    case "dm_error": {
      console.warn("DM error:", msg.error);
      break;
    }

    case "thread_bump": {
      const idx = threads.findIndex(t => t.other_user_id === msg.other_user_id);
      if (idx !== -1) {
        threads[idx].last_message_body = msg.last_message_body;
        threads[idx].last_message_at = msg.last_message_at;
        threads[idx].last_message_sender = msg.last_message_sender;
      }
      renderUsers(sortUsers(toThreadUserListItems()));
      break;
    }

    case "presence": {
      const { user_id, online } = msg;

      threads.forEach(t => {
        if (t.other_user_id === user_id) {
          t.online = online;
        }
      });

      suggestedUsers.forEach(u => {
        if (u.user_id === user_id) {
          u.online = online;
        }
      });

      renderUsers(sortUsers(toThreadUserListItems()));
      break;
    }


    case "presence_snapshot": {
      const onlineSet = new Set(msg.online_ids || []);

      threads.forEach(t => {
        t.online = onlineSet.has(t.other_user_id);
      });

      suggestedUsers.forEach(u => {
        u.online = onlineSet.has(u.user_id);
      });

      renderUsers(sortUsers(toThreadUserListItems()));
      break;
    }

    default:
      // ignore unknown types
      break;
  }
}

// -----------------------------
// Sending
// -----------------------------
function genClientMsgID() {
  if (globalThis.crypto && typeof globalThis.crypto.randomUUID === "function") {
    return globalThis.crypto.randomUUID();
  }

  // fallback: RFC4122-ish v4 UUID using crypto.getRandomValues
  if (globalThis.crypto && typeof globalThis.crypto.getRandomValues === "function") {
    const b = new Uint8Array(16);
    globalThis.crypto.getRandomValues(b);

    // v4
    b[6] = (b[6] & 0x0f) | 0x40;
    // variant
    b[8] = (b[8] & 0x3f) | 0x80;

    const hex = [...b].map(x => x.toString(16).padStart(2, "0")).join("");
    return (
      hex.slice(0, 8) + "-" +
      hex.slice(8, 12) + "-" +
      hex.slice(12, 16) + "-" +
      hex.slice(16, 20) + "-" +
      hex.slice(20)
    );
  }

  // last resort (still unique enough for client_msg_id)
  return "m_" + Date.now().toString(36) + "_" + Math.random().toString(36).slice(2);
}

function sendActiveMessage() {
  if (!activeChatUser) return;

  const input = document.getElementById("chat-input");
  const body = (input.value || "").trim();
  if (!body) return;

  sendDM(activeChatUser.id, body);
  input.value = "";
}

function sendDM(toUserID, body) {
  const clientMsgID = genClientMsgID();

  const payload = {
    type: "dm_send",
    to_user_id: toUserID,
    body,
    client_msg_id: clientMsgID
  };

  pendingMessages.set(clientMsgID, payload);
  wsManager.send("dm", payload);
}

// -----------------------------
// Init
// -----------------------------
function connect() {
  wsManager.connect("dm", "/ws/dm");
  wsManager.subscribe("dm", handleWsMessage);
}

document.addEventListener("DOMContentLoaded", async () => {
  const loggedIn = document.body.dataset.showLogin !== "1";
  if (!loggedIn) return;

  createChatSidebar();
  await loadThreads();
  connect();

  window.addEventListener("scroll", scheduleChatSidebarStopUpdate, { passive: true });
  window.addEventListener("resize", scheduleChatSidebarStopUpdate);
  scheduleChatSidebarStopUpdate();
});



// Add this helper function to clear unread when un-minimizing:
function clearUnreadForActiveChat() {
  if (!activeChatUser) return;
  const t = threads.find(t => t.other_user_id === activeChatUser.id);
  if (t) {
    t.unread_count = 0;
  }
  renderUsers(sortUsers(toThreadUserListItems()));
}

