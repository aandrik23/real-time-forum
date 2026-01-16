// ws.js
import { authFetch } from "./auth.js";
import { isAnonymous } from "./utils.js";

const sockets = new Map(); // channel -> socket state

const RECONNECT_BASE_MS = 500;
const RECONNECT_MAX_MS = 10000;
const MAX_RECONNECT_ATTEMPTS = 6;

function wsURL(path) {
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${location.host}${path}`;
}

function createState(channel, path) {
  return {
    channel,
    path,
    ws: null,
    reconnectTimer: null,
    reconnectAttempt: 0,
    manualClose: false,
    subscribers: new Set()
  };
}

function connect(channel, path) {
  if (isAnonymous()) return;

  let state = sockets.get(channel);
  if (!state) {
    state = createState(channel, path);
    sockets.set(channel, state);
  }

  if (state.ws &&
      (state.ws.readyState === WebSocket.OPEN ||
       state.ws.readyState === WebSocket.CONNECTING)) {
    return;
  }

  state.manualClose = false;
  state.ws = new WebSocket(wsURL(path));

  state.ws.addEventListener("open", () => {
    state.reconnectAttempt = 0;
    if (state.reconnectTimer) {
      clearTimeout(state.reconnectTimer);
      state.reconnectTimer = null;
    }
  });

  state.ws.addEventListener("message", ev => {
    for (const fn of state.subscribers) {
      fn(ev);
    }
  });

  state.ws.addEventListener("close", e => {
    if (e.code === 1008) return;
    if (!state.manualClose) scheduleReconnect(state);
  });
}

function scheduleReconnect(state) {
  if (state.reconnectTimer) return;
  if (state.reconnectAttempt >= MAX_RECONNECT_ATTEMPTS) return;

  const exp = Math.min(
    RECONNECT_MAX_MS,
    RECONNECT_BASE_MS * (2 ** state.reconnectAttempt++)
  );

  const jitter = exp * (Math.random() * 0.4 - 0.2);
  const delay = Math.max(250, Math.floor(exp + jitter));

  state.reconnectTimer = setTimeout(async () => {
    state.reconnectTimer = null;
    try {
      const res = await authFetch("/api/auth/ping");
      if (!res.ok) throw new Error();
      connect(state.channel, state.path);
    } catch {
      // auth invalid, do nothing
       console.warn("WS reconnect aborted: auth invalid"); //debug
    }
  }, delay);
}

function subscribe(channel, handler) {
  const state = sockets.get(channel);
  if (!state) throw new Error(`WS channel not connected: ${channel}`);
  state.subscribers.add(handler);
}

function send(channel, payload) {
  const state = sockets.get(channel);
  if (!state || !state.ws || state.ws.readyState !== WebSocket.OPEN) return;
  state.ws.send(JSON.stringify(payload));
}

function disconnect(channel) {
  const state = sockets.get(channel);
  if (!state) return;

  state.manualClose = true;
  if (state.reconnectTimer) {
    clearTimeout(state.reconnectTimer);
    state.reconnectTimer = null;
  }
  if (state.ws) {
    state.ws.close();
    state.ws = null;
  }
}

function resetAll() {
  for (const channel of sockets.keys()) {
    disconnect(channel);
  }
  sockets.clear();
}

// passive recovery
window.addEventListener("online", () => {
  for (const state of sockets.values()) {
    connect(state.channel, state.path);
  }
});

document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "visible") {
    for (const state of sockets.values()) {
      connect(state.channel, state.path);
    }
  }
});

export default {
  connect,
  subscribe,
  send,
  disconnect,
  resetAll
};
