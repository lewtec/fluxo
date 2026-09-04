const THEME_KEY = "fluxo-theme";

function currentTheme() {
  return document.documentElement.getAttribute("data-theme") === "listing" ? "listing" : "bay";
}

function applyTheme(theme) {
  const next = theme === "listing" ? "listing" : "bay";
  document.documentElement.setAttribute("data-theme", next);
  try {
    localStorage.setItem(THEME_KEY, next);
  } catch (e) {
    // private mode / blocked storage
  }
}

function parseRoot(html) {
  const tpl = document.createElement("template");
  tpl.innerHTML = html.trim();
  return tpl.content.firstElementChild;
}

function swapHTML(id, html) {
  const el = document.getElementById(id);
  if (!el) return;
  const next = parseRoot(html);
  if (next) el.replaceWith(next);
}

// Live regions inside the detail page. The collapse checkboxes stay in the
// document so opening Files/Trackers survives speed/progress ticks.
const DETAIL_PATCH = [
  "torrent-detail-actions",
  "torrent-detail-summary",
  "torrent-files-title",
  "torrent-files-body",
  "torrent-trackers-title",
  "torrent-trackers-body",
];

function patchDetail(html) {
  const el = document.getElementById("torrent-detail");
  if (!el) return;
  const next = parseRoot(html);
  if (!next) return;
  if (next.dataset.id && el.dataset.id && next.dataset.id !== el.dataset.id) return;
  for (const id of DETAIL_PATCH) {
    const cur = document.getElementById(id);
    const incoming = next.querySelector("#" + id);
    if (cur && incoming) cur.replaceWith(incoming);
  }
}

function connectEvents() {
  const es = new EventSource("/events");
  es.addEventListener("stats", (e) => swapHTML("header-stats", e.data));
  es.addEventListener("list", (e) => swapHTML("torrent-list", e.data));
  es.addEventListener("detail", (e) => patchDetail(e.data));
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
  box.checked = currentTheme() === "listing";
  box.addEventListener("change", () => {
    applyTheme(box.checked ? "listing" : "bay");
  });
}

bindTheme();
connectEvents();
