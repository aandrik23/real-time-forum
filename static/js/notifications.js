import wsManager from "./ws.js";
import { onAppReset } from "./appReset.js";
import { isAnonymous } from "./utils.js";
// -----------------------------
// App reset
// -----------------------------
onAppReset(() => {
  notifications = [];
  if (notifList) notifList.innerHTML = "";
  if (notifBadge) notifBadge.textContent = "0";
  if (notifDropdown) notifDropdown.style.display = "none";
});

// Elements
const notifBtn = document.querySelector(".notif-btn");
const notifDropdown = document.getElementById("notif-dropdown");
const notifList = document.getElementById("notif-list");
const notifBadge = notifBtn?.querySelector(".badge");


let notifications = [];

// Toggle dropdown on button click
notifBtn.addEventListener("click", () => {
    const isOpen = notifDropdown.style.display !== "none";

    if (!isOpen) {
        notifDropdown.style.display = "block";
        notifications.forEach(n => n.read = true);
        notifBadge.textContent = "0";
    } else {
        notifDropdown.style.display = "none";
    }
});

function handleNotificationMessage(ev) {
  let data;
  try {
    data = JSON.parse(ev.data);
  } catch {
    return;
  }

  // Ignore the "connected" message
  if (data.type === "connected") return;

  notifications.unshift({
    ...data,
    read: false
  });

  const unreadCount = notifications.filter(n => !n.read).length;
  notifBadge.textContent = unreadCount;

  const li = document.createElement("li");

  let text = "New notification";
  if (data.type === "comment") {
    text = data.payload?.message || "Someone commented on your post";
  } else if (data.type === "like") {
    text = data.payload?.message || "Someone liked your post";
  }

  li.textContent = text;
  notifList.prepend(li);
}

function connectNotifications() {
    if (isAnonymous()) return;

    wsManager.connect("notifications", "/ws/notifications");
    wsManager.subscribe("notifications", handleNotificationMessage);
}

document.addEventListener("DOMContentLoaded", () => {
  connectNotifications();
});
