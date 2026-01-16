// Elements
const notifBtn = document.getElementById("notif-btn");
const notifDropdown = document.getElementById("notif-dropdown");
const notifList = document.getElementById("notif-list");
const notifBadge = document.getElementById("notif-badge");

let notifications = [];
let ws = null;

// Toggle dropdown on button click
notifBtn.addEventListener("click", () => {
    if (notifDropdown.style.display === "none") {
        notifDropdown.style.display = "block";

        // Optional: mark all as read when opened
        notifications.forEach(n => n.read = true);
        notifBadge.textContent = "0";

    } else {
        notifDropdown.style.display = "none";
    }
});

// Connect to WS
function wsURL(path) {
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${location.host}${path}`;
}

ws = new WebSocket(wsURL("/ws/notifications"));

ws.onopen = () => {
    console.log("Notification WS connected");
} 

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);

    // Ignore the "connected" message
    if (data.type === "connected") return;

    // Push new notification to array
    notifications.unshift(data);

    // Update badge count for unread notifications
    const unreadCount = notifications.filter(n => !n.read).length;
    notifBadge.textContent = unreadCount;

    // Render notification in dropdown
    const li = document.createElement("li");

    // Display a descriptive message
    let text = "New notification";
    if (data.type === "comment") {
        text = data.payload?.message || "Someone commented on your post";
    } else if (data.type === "like") {
        text = data.payload?.message || "Someone liked your post";
    }

    li.textContent = text;
    notifList.prepend(li);
};
