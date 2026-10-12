// bururu.dev: the menus and the search. The pages work without this
// file: the menus still open and close on a click, they only lose the
// place on the page when you switch. The search needs it, so it stays
// hidden until this file shows it.
"use strict";

(() => {
  const menus = [...document.querySelectorAll("details.menu")];
  const box = document.querySelector(".search");
  const input = box && box.querySelector("input");
  const list = box && box.querySelector(".results");

  // The part of an address after /<version>/<language>/: the page.
  const pageOf = (p) => p.split("/").slice(3).join("/");

  // The version and language menus keep the place on the page: a link to
  // the same page gets the heading in the address (#...), since headings
  // have the same ids in every language.
  function keepPlace(m) {
    if (!m.matches(".menu-version, .menu-lang")) return;
    const here = pageOf(location.pathname);
    for (const a of m.querySelectorAll(".pop a")) {
      const url = new URL(a.href);
      url.hash = pageOf(url.pathname) === here ? location.hash : "";
      a.href = url.href;
    }
  }

  // One menu open at a time; a click elsewhere or Escape closes it.
  for (const m of menus) {
    m.addEventListener("toggle", () => {
      if (!m.open) return;
      for (const other of menus) if (other !== m) other.open = false;
      hideResults();
      keepPlace(m);
    });
  }
  document.addEventListener("click", (e) => {
    for (const m of menus) if (m.open && !m.contains(e.target)) m.open = false;
    if (box && !box.contains(e.target)) hideResults();
  });
  document.addEventListener("keydown", (e) => {
    if (e.key !== "Escape") return;
    const open = menus.find((m) => m.open);
    if (open) {
      open.open = false;
      open.querySelector("summary").focus();
    }
  });

  if (!box) return;
  box.hidden = false;

  // The index: every section of every page as one entry. It loads the
  // first time the field gets the focus.
  let entries = null;
  let loading = null;
  function load() {
    if (!loading) {
      loading = fetch(box.dataset.index)
        .then((r) => (r.ok ? r.json() : Promise.reject(new Error(r.status))))
        .then((data) => {
          entries = [];
          for (const p of data.pages) {
            for (const [heading, id, text] of p.s) {
              entries.push({
                page: p.t,
                heading,
                url: box.dataset.base + p.u + (id ? "#" + id : ""),
                text,
                head: (p.t + " " + heading).toLowerCase(),
                all: (p.t + " " + heading + " " + text).toLowerCase(),
              });
            }
          }
        })
        .catch(() => {
          loading = null;
        });
    }
    return loading || Promise.resolve();
  }

  function hideResults() {
    if (!list) return;
    list.hidden = true;
    input.setAttribute("aria-expanded", "false");
  }

  function count(hay, term) {
    let n = 0;
    for (let i = hay.indexOf(term); i >= 0 && n < 5; i = hay.indexOf(term, i + term.length)) n++;
    return n;
  }

  // A piece of the text around the first word found.
  function snippet(text, terms) {
    const lower = text.toLowerCase();
    let at = -1;
    for (const t of terms) {
      const i = lower.indexOf(t);
      if (i >= 0 && (at < 0 || i < at)) at = i;
    }
    const start = Math.max(0, at - 40);
    const end = Math.min(text.length, Math.max(at, 0) + 110);
    return (start > 0 ? "..." : "") + text.slice(start, end) + (end < text.length ? "..." : "");
  }

  // Put text into el with the words found marked.
  function marked(el, text, terms) {
    const lower = text.toLowerCase();
    let pos = 0;
    while (pos < text.length) {
      let at = -1;
      let len = 0;
      for (const t of terms) {
        const i = lower.indexOf(t, pos);
        if (i >= 0 && (at < 0 || i < at)) {
          at = i;
          len = t.length;
        }
      }
      if (at < 0) break;
      el.append(text.slice(pos, at));
      const m = document.createElement("mark");
      m.textContent = text.slice(at, at + len);
      el.append(m);
      pos = at + len;
    }
    el.append(text.slice(pos));
  }

  function show() {
    const q = input.value.trim().toLowerCase();
    list.replaceChildren();
    if (!q || !entries) {
      hideResults();
      return;
    }
    const terms = q.split(/\s+/);
    const hits = [];
    for (const e of entries) {
      if (!terms.every((t) => e.all.includes(t))) continue;
      let score = 0;
      for (const t of terms) score += (e.head.includes(t) ? 10 : 0) + count(e.all, t);
      hits.push([score, e]);
    }
    hits.sort((a, b) => b[0] - a[0]);
    if (hits.length === 0) {
      const p = document.createElement("p");
      p.className = "empty";
      p.textContent = list.dataset.empty;
      list.append(p);
    }
    for (const [, e] of hits.slice(0, 12)) {
      const a = document.createElement("a");
      a.href = e.url;
      const title = document.createElement("span");
      title.className = "r-title";
      marked(title, e.heading || e.page, terms);
      a.append(title);
      if (e.heading) {
        const page = document.createElement("span");
        page.className = "r-page";
        page.textContent = e.page;
        a.append(page);
      }
      const snip = document.createElement("span");
      snip.className = "r-snip";
      marked(snip, snippet(e.text, terms), terms);
      a.append(snip);
      a.addEventListener("click", hideResults);
      list.append(a);
    }
    list.hidden = false;
    input.setAttribute("aria-expanded", "true");
  }

  input.addEventListener("focus", () => {
    load();
    if (input.value.trim()) show();
  });
  input.addEventListener("input", () => load().then(show));

  // Arrow keys move through the results, Enter opens the first one.
  function links() {
    return [...list.querySelectorAll("a")];
  }
  input.addEventListener("keydown", (e) => {
    if (e.key === "ArrowDown") {
      const first = links()[0];
      if (first) {
        e.preventDefault();
        first.focus();
      }
    } else if (e.key === "Enter") {
      const first = links()[0];
      if (first) {
        e.preventDefault();
        first.click();
      }
    } else if (e.key === "Escape") {
      input.value = "";
      hideResults();
    }
  });
  list.addEventListener("keydown", (e) => {
    const all = links();
    const i = all.indexOf(document.activeElement);
    if (e.key === "ArrowDown" && i < all.length - 1) {
      e.preventDefault();
      all[i + 1].focus();
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      (i > 0 ? all[i - 1] : input).focus();
    } else if (e.key === "Escape") {
      hideResults();
      input.focus();
    }
  });
  box.addEventListener("focusout", (e) => {
    if (!box.contains(e.relatedTarget)) hideResults();
  });
})();
