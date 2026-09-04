const THEME_KEY = "fluxo-theme";

function currentTheme() {
  return document.documentElement.getAttribute("data-theme") === "light" ? "light" : "dark";
}

function persistTheme(theme) {
  try {
    localStorage.setItem(THEME_KEY, theme);
  } catch (e) {
    // private mode / blocked storage
  }
}

function swapHTML(id, html) {
  const el = document.getElementById(id);
  if (!el) return;
  const tpl = document.createElement("template");
  tpl.innerHTML = html.trim();
  const next = tpl.content.firstElementChild;
  if (next) el.replaceWith(next);
}

function connectEvents() {
  const es = new EventSource("/events");
  es.addEventListener("stats", (e) => swapHTML("header-stats", e.data));
  es.addEventListener("list", (e) => swapHTML("torrent-list", e.data));
  es.addEventListener("detail", (e) => {
    const el = document.getElementById("torrent-detail");
    if (!el) return;
    const tpl = document.createElement("template");
    tpl.innerHTML = e.data.trim();
    const next = tpl.content.firstElementChild;
    if (!next) return;
    if (next.dataset.id && el.dataset.id && next.dataset.id !== el.dataset.id) return;
    el.replaceWith(next);
  });
  es.addEventListener("removed", (e) => {
    const el = document.getElementById("torrent-detail");
    if (el && el.dataset.id === e.data.trim()) {
      window.location.assign("/");
    }
  });
}

function bindTheme() {
  const box = document.getElementById("theme-toggle");
  if (!box) return;
  box.checked = currentTheme() === "light";
  box.addEventListener("change", () => {
    persistTheme(box.checked ? "light" : "dark");
  });
}

bindTheme();
connectEvents();
