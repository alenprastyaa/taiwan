const sidebarOpenClass = "sidebar-open";
let chatSocket;

function closeSidebar() {
  document.body.classList.remove(sidebarOpenClass);
}

function bindShell() {
  document.querySelectorAll("[data-sidebar-toggle]").forEach((button) => {
    button.addEventListener("click", () => {
      document.body.classList.toggle(sidebarOpenClass);
    });
  });

  document.querySelectorAll("[data-sidebar-close], .nav-link").forEach((element) => {
    element.addEventListener("click", closeSidebar);
  });
}

function closeChatSocket() {
  if (chatSocket) {
    chatSocket.close();
    chatSocket = undefined;
  }
}

function appendChatMessage(chat, message) {
  const windowElement = chat.querySelector("[data-chat-window]");
  if (!windowElement || windowElement.querySelector(`[data-message-id="${message.id}"]`)) {
    return;
  }

  const row = document.createElement("div");
  row.className = `message ${message.senderId === chat.dataset.userId ? "self" : "other"}`;
  row.dataset.messageId = message.id;

  const sender = document.createElement("strong");
  sender.textContent = message.senderName;
  const body = document.createElement("span");
  body.textContent = message.body;

  row.append(sender, body);
  windowElement.append(row);
  windowElement.scrollTop = windowElement.scrollHeight;
}

function bindChat() {
  closeChatSocket();

  const chat = document.querySelector("[data-chat]");
  if (!chat) {
    return;
  }

  const conversationId = chat.dataset.conversationId;
  const form = chat.querySelector("[data-chat-form]");
  const input = form?.querySelector("input[name='body']");
  if (!conversationId || !form || !input) {
    return;
  }

  const scheme = window.location.protocol === "https:" ? "wss" : "ws";
  chatSocket = new WebSocket(`${scheme}://${window.location.host}/ws/chat/${encodeURIComponent(conversationId)}`);

  chatSocket.addEventListener("message", (event) => {
    try {
      const payload = JSON.parse(event.data);
      if (payload.type === "message" && payload.message) {
        appendChatMessage(chat, payload.message);
      }
    } catch {
      // Ignore malformed websocket payloads from interrupted local sessions.
    }
  });

  form.addEventListener("submit", (event) => {
    if (!chatSocket || chatSocket.readyState !== WebSocket.OPEN) {
      return;
    }

    event.preventDefault();
    const body = input.value.trim();
    if (!body) {
      return;
    }
    chatSocket.send(JSON.stringify({ body }));
    input.value = "";
  });
}

async function loadApp(url, pushState) {
  const response = await fetch(url, {
    headers: {
      "HX-Request": "true",
      "X-Requested-With": "fetch"
    }
  });

  if (!response.ok) {
    window.location.href = url;
    return;
  }

  const html = await response.text();
  const app = document.querySelector("#app");
  if (!app) {
    window.location.href = url;
    return;
  }

  app.outerHTML = html;
  closeSidebar();
  bindShell();
  bindChat();
  syncTitle();

  if (pushState) {
    history.pushState({ url }, "", url);
  }

  window.scrollTo({ top: 0, left: 0 });
}

function bindNavigation() {
  document.addEventListener("click", (event) => {
    const link = event.target.closest("a[hx-get]");
    if (!link) {
      return;
    }

    if (event.defaultPrevented || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
      return;
    }

    const target = link.getAttribute("target");
    if (target && target !== "_self") {
      return;
    }

    const route = link.getAttribute("hx-get") || link.getAttribute("href");
    if (!route) {
      return;
    }

    const url = new URL(route, window.location.href);
    if (url.origin !== window.location.origin) {
      return;
    }

    event.preventDefault();
    loadApp(url.pathname + url.search + url.hash, true).catch(() => {
      window.location.href = url.href;
    });
  });

  window.addEventListener("popstate", () => {
    loadApp(window.location.pathname + window.location.search + window.location.hash, false).catch(() => {
      window.location.reload();
    });
  });
}

function syncTitle() {
  const app = document.querySelector("#app");
  const title = app?.getAttribute("data-page-title");
  if (title) {
    document.title = title;
  }
}

async function registerServiceWorker() {
  if (!("serviceWorker" in navigator)) {
    return;
  }

  try {
    await navigator.serviceWorker.register("/sw.js");
  } catch {
    // The app still works when service worker registration is blocked locally.
  }
}

document.addEventListener("DOMContentLoaded", () => {
  bindShell();
  bindNavigation();
  bindChat();
  syncTitle();
  registerServiceWorker();
});

window.addEventListener("beforeunload", closeChatSocket);
