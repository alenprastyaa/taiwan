const sidebarOpenClass = "sidebar-open";
let chatSocket;

function bindAutoSubmitSelects() {
  // Delegated so it keeps working after htmx swaps #app (pipeline stage-move selects).
  document.addEventListener("change", (event) => {
    const select = event.target.closest("select[data-autosubmit]");
    if (select && select.form) {
      select.form.requestSubmit();
    }
  });
}

function bindConfirmForms() {
  // Delegated confirm() guard for destructive actions (e.g. deleting a pipeline stage).
  document.addEventListener("submit", (event) => {
    const form = event.target.closest("form[data-confirm]");
    if (form && !window.confirm(form.dataset.confirm)) {
      event.preventDefault();
    }
  });
}

function bindCopyButtons() {
  // Delegated clipboard copy for text-template "Salin Teks" buttons.
  document.addEventListener("click", (event) => {
    const button = event.target.closest("[data-copy-text]");
    if (!button) return;
    navigator.clipboard.writeText(button.dataset.copyText).then(() => {
      const original = button.textContent;
      button.textContent = "Disalin!";
      setTimeout(() => {
        button.textContent = original;
      }, 1500);
    });
  });
}

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

function bindDemoAccounts() {
  document.querySelectorAll(".demo-account").forEach((button) => {
    button.addEventListener("click", () => {
      const form = button.closest(".auth-panel")?.querySelector(".auth-form");
      const usernameInput = form?.querySelector('[name="username"]');
      const passwordInput = form?.querySelector('[name="password"]');
      if (!usernameInput || !passwordInput) {
        return;
      }
      usernameInput.value = button.dataset.demoUsername;
      passwordInput.value = button.dataset.demoPassword;
      passwordInput.focus();
    });
  });
}

function bindRejectModal() {
  // Delegated handlers, attached once — survive htmx #app swaps.
  document.addEventListener("click", (event) => {
    const trigger = event.target.closest("[data-reject-url]");
    if (trigger) {
      const modal = document.querySelector("[data-reject-modal]");
      if (!modal) {
        return;
      }
      const form = modal.querySelector("[data-reject-form]");
      const textarea = form.querySelector('textarea[name="reason"]');
      form.setAttribute("action", trigger.dataset.rejectUrl);
      modal.querySelector("[data-reject-target]").textContent = trigger.dataset.rejectName || "";
      textarea.value = "";
      modal.removeAttribute("hidden");
      textarea.focus();
      return;
    }
    const openModal = document.querySelector("[data-reject-modal]:not([hidden])");
    if (openModal && (event.target.closest("[data-reject-cancel]") || event.target === openModal)) {
      openModal.setAttribute("hidden", "");
    }
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      const openModal = document.querySelector("[data-reject-modal]:not([hidden])");
      if (openModal) {
        openModal.setAttribute("hidden", "");
      }
    }
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
  bindDemoAccounts();
  bindRejectModal();
  bindAutoSubmitSelects();
  bindConfirmForms();
  bindCopyButtons();
  syncTitle();
  registerServiceWorker();
});

window.addEventListener("beforeunload", closeChatSocket);
