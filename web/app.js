const statusEl = document.getElementById("status");
const preview = document.getElementById("preview");
const colors = ["W", "U", "B", "R", "G"];
const rarityNames = ["common", "uncommon", "rare", "mythic"];
let selected = new Set();
let pickedRarity = new Set();
let priceSort = "";
let deckId = 0;
let face = {};

function say(text) { statusEl.textContent = text || ""; }

async function api(method, path, body) {
  const opt = { method, headers: {} };
  if (body !== undefined) {
    opt.headers["Content-Type"] = "application/json";
    opt.body = JSON.stringify(body);
  }
  const res = await fetch(path, opt);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || res.statusText);
  const worth = res.headers.get("X-Collection-USD");
  if (worth) showWorth(worth);
  return data;
}

function showWorth(raw) {
  const n = Number(raw);
  if (!Number.isFinite(n)) return;
  document.getElementById("worth-value").textContent = n.toLocaleString("en-US", { style: "currency", currency: "USD" });
}

function imgURL(path) {
  if (!path) return "";
  if (path.startsWith("http")) return path;
  return "/images/" + path;
}

function pips(cost) {
  if (!cost) return "";
  return cost.replace(/\{([^}]+)\}/g, (_, raw) => {
    const sym = raw.toUpperCase();
    const cls = "WUBRG".includes(sym) ? sym.toLowerCase() : "c";
    return `<i class="pip ${cls}">${sym.length > 2 ? sym : sym}</i>`;
  });
}

function sectionOf(typeLine) {
  const t = (typeLine || "").toLowerCase();
  if (t.includes("land")) return "Lands";
  if (t.includes("creature")) return "Creatures";
  return "Spells";
}

function cellHTML(c, i, mode) {
  const usedElsewhere = mode === "deck" && c.in_other_built > 0;
  const locked = usedElsewhere && (c.owned - c.in_other_built) <= 0;
  const label = usedElsewhere ? `In deck: ${c.other_decks}` : "";
  const src = imgURL(face[key(c)] || c.front_image);
  const art = src
    ? `<img alt="" src="${src}">`
    : `<div class="plate">${escapeHTML(c.name || "Unlinked printing")}</div>`;
  const count = mode === "deck" ? c.in_this : c.owned;
  return `<button class="cell${locked ? " locked" : ""}" type="button" data-i="${i}">
    <div class="art">${art}
      ${count ? `<span class="badge">×${count}</span>` : ""}
      ${label ? `<span class="veil">${escapeHTML(label)}</span>` : ""}
    </div>
    <figcaption>
      <strong>${escapeHTML(c.name || c.set + " " + c.number)}</strong>
      <em>${escapeHTML(c.set_name || c.set || "Choose a printing")}</em>
      <div class="stat">${escapeHTML([c.rarity, c.foil ? "foil" : "nonfoil", c.price_usd ? "$" + c.price_usd : "", "×" + c.owned].filter(Boolean).join(" · "))}</div>
    </figcaption>
  </button>`;
}

function key(c) { return `${c.set}|${c.number}|${c.foil ? 1 : 0}`; }
function escapeHTML(s) { return String(s).replace(/[&<>"]/g, ch => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[ch])); }

function show(id) {
  preview.style.display = "none";
  document.querySelectorAll(".view").forEach(v => v.classList.remove("on"));
  document.getElementById(id).classList.add("on");
  document.querySelectorAll(".topbar nav button").forEach(b => b.classList.toggle("on", b.dataset.view === id || (id === "editor" && b.dataset.view === "decks")));
}

function colorButtons(host) {
  host.innerHTML = colors.map(c => `<button type="button" class="pip ${c.toLowerCase()}" data-color="${c}">${c}</button>`).join("");
  host.querySelectorAll("button").forEach(btn => btn.addEventListener("click", () => {
    const c = btn.dataset.color;
    if (selected.has(c)) selected.delete(c); else selected.add(c);
    btn.classList.toggle("on", selected.has(c));
    if (document.getElementById("editor").classList.contains("on")) renderPool(lastPool);
    else renderOwned(lastCards);
  }));
}

let lastCards = [];
let lastPool = [];
let lastEntries = [];

function matchesColor(c) {
  if (selected.size === 0) return true;
  for (const col of selected) if ((c.colors || "").includes(col)) return true;
  return false;
}

function matchesRarity(c) {
  if (pickedRarity.size === 0) return true;
  return pickedRarity.has((c.rarity || "").toLowerCase());
}

function byPrice(cards) {
  if (priceSort !== "asc" && priceSort !== "desc") return cards;
  const priced = [];
  const rest = [];
  for (const c of cards) {
    if (c.price_usd === "" || !Number.isFinite(Number(c.price_usd))) rest.push(c);
    else priced.push(c);
  }
  priced.sort((a, b) => {
    const d = Number(a.price_usd) - Number(b.price_usd);
    if (d !== 0) return priceSort === "asc" ? d : -d;
    return String(a.name).localeCompare(String(b.name));
  });
  return priced.concat(rest);
}

function visibleOwned(cards) {
  return byPrice(cards.filter(c => matchesColor(c) && matchesRarity(c)));
}

function renderOwned(cards) {
  lastCards = cards;
  const shown = visibleOwned(cards);
  const empty = cards.length
    ? "No cards match these filters."
    : "No cards yet. Look up a printing, or import a kitchen archive.";
  document.getElementById("owned").innerHTML = shown.map((c, i) => cellHTML(c, i)).join("") || `<p class="quiet">${empty}</p>`;
  document.getElementById("owned").querySelectorAll(".cell").forEach(el => {
    el.addEventListener("click", () => previewCard(shown[+el.dataset.i]));
  });
}

async function loadCards() {
  const q = document.getElementById("q").value.trim();
  const cards = await api("GET", "/api/cards?q=" + encodeURIComponent(q));
  renderOwned(cards);
}

async function loadDecks() {
  const decks = await api("GET", "/api/decks");
  const grid = document.getElementById("deckgrid");
  grid.innerHTML = decks.map(d => `<button class="box" type="button" data-id="${d.id}">
      ${d.cover ? `<img alt="" src="${imgURL(d.cover)}">` : `<div class="cover"></div>`}
      <div><strong>${escapeHTML(d.name || "Untitled")}</strong><span>${d.status} · ${d.cards} cards</span></div>
    </button>`).join("") + `<button class="addbox" type="button" id="newdeck" aria-label="New deck">+</button>`;
  grid.querySelectorAll(".box").forEach(el => el.addEventListener("click", () => openDeck(+el.dataset.id)));
  document.getElementById("newdeck").addEventListener("click", async () => {
    const created = await api("POST", "/api/decks", { name: "New deck", description: "", status: "draft" });
    openDeck(created.id);
  });
}

async function openDeck(id) {
  deckId = id;
  show("editor");
  await refreshDeck();
}

async function refreshDeck() {
  const q = document.getElementById("dq").value.trim();
  const [detail, pool] = await Promise.all([
    api("GET", "/api/decks/" + deckId),
    api("GET", "/api/pool?deck_id=" + deckId + "&q=" + encodeURIComponent(q)),
  ]);
  lastEntries = detail.entries;
  lastPool = pool;
  renderPool(pool);
  renderList(detail.deck, detail.entries);
}

function renderPool(pool) {
  const shown = pool.filter(matchesColor);
  document.getElementById("pool").innerHTML = shown.map((c, i) => cellHTML(c, i, "deck")).join("");
  document.getElementById("pool").querySelectorAll(".cell").forEach(el => {
    const c = shown[+el.dataset.i];
    el.addEventListener("click", () => addCopy(c));
  });
}

function renderList(deck, entries) {
  const total = entries.reduce((n, e) => n + e.qty, 0);
  const groups = { Creatures: [], Spells: [], Lands: [] };
  for (const e of entries) groups[sectionOf(e.type_line)].push(e);
  let rows = "";
  for (const name of ["Creatures", "Spells", "Lands"]) {
    const list = groups[name];
    if (!list.length) continue;
    rows += `<div class="sec">${name} ${list.reduce((n, e) => n + e.qty, 0)}</div>`;
    for (const e of list) {
      rows += `<button class="drow" type="button" data-set="${escapeHTML(e.set)}" data-number="${escapeHTML(e.number)}" data-foil="${e.foil ? 1 : 0}">
        <span class="n">${e.qty}</span><span class="nm">${escapeHTML(e.name || e.set + " " + e.number)}</span><span>${pips(e.mana_cost)}</span>
      </button>`;
    }
  }
  document.getElementById("list").innerHTML = `
    <input class="dname" id="dname" value="${escapeHTML(deck.name)}" aria-label="Deck name">
    <input class="ddesc" id="ddesc" value="${escapeHTML(deck.description)}" aria-label="Description">
    <div class="modes">
      <button type="button" id="mode-draft" class="${deck.status === "draft" ? "on" : ""}">Draft</button>
      <button type="button" id="mode-built" class="${deck.status === "built" ? "on" : ""}">Built</button>
    </div>
    <div class="count">${total}<small> / 60</small></div>
    ${rows}`;
  document.getElementById("dname").addEventListener("change", saveMeta);
  document.getElementById("ddesc").addEventListener("change", saveMeta);
  document.getElementById("mode-draft").addEventListener("click", () => saveStatus("draft"));
  document.getElementById("mode-built").addEventListener("click", () => saveStatus("built"));
  document.getElementById("list").querySelectorAll(".drow").forEach(el => {
    el.addEventListener("click", () => removeCopy(el.dataset.set, el.dataset.number, el.dataset.foil === "1"));
  });
}

async function saveMeta() {
  await api("PATCH", "/api/decks/" + deckId, {
    name: document.getElementById("dname").value,
    description: document.getElementById("ddesc").value,
  });
}

async function saveStatus(status) {
  try {
    await api("PATCH", "/api/decks/" + deckId, { status });
    say(status === "built" ? "Deck is built. Its copies are locked." : "Draft. Copies stay free for other decks.");
    await refreshDeck();
  } catch (err) {
    say(err.message);
  }
}

async function addCopy(c) {
  const free = c.owned - c.in_other_built - c.in_this;
  if (free <= 0 && c.other_decks) {
    say("In deck: " + c.other_decks);
    return;
  }
  if (free <= 0) {
    say("No copies left.");
    return;
  }
  try {
    await api("POST", "/api/decks/" + deckId + "/entries", { set: c.set, number: c.number, foil: c.foil, qty: 1 });
    say("");
    await refreshDeck();
  } catch (err) {
    say(err.message);
  }
}

async function removeCopy(set, number, foil) {
  try {
    await api("POST", "/api/decks/" + deckId + "/entries", { set, number, foil, qty: -1 });
    await refreshDeck();
  } catch (err) {
    say(err.message);
  }
}

function previewCard(c) {
  if (!c) return;
  const src = imgURL(face[key(c)] || c.front_image);
  if (!src) return;
  preview.style.display = "block";
  preview.innerHTML = `<img id="pv" alt="" src="${src}" data-front="${imgURL(c.front_image)}" data-back="${imgURL(c.back_image)}">` +
    (c.back_image ? `<button type="button" id="flip">Flip</button>` : "");
}

function placePreview(anchor, card) {
  if (!card || !card.front_image) {
    preview.style.display = "none";
    return;
  }
  const r = anchor.getBoundingClientRect();
  preview.style.paddingLeft = "0";
  preview.style.paddingRight = "0";
  let left = r.right;
  if (r.right + 260 > window.innerWidth) left = Math.max(8, r.left - 260);
  preview.style.left = left + "px";
  previewCard(card);
  const height = preview.offsetHeight;
  let top = r.top;
  if (top + height > window.innerHeight - 8) top = Math.max(8, window.innerHeight - 8 - height);
  preview.style.top = top + "px";
}

document.body.addEventListener("mouseover", (e) => {
  const row = e.target.closest(".drow");
  if (row) {
    const card = lastEntries.find(en => en.set === row.dataset.set && en.number === row.dataset.number && !!en.foil === (row.dataset.foil === "1"));
    placePreview(row, card);
    return;
  }
  const cell = e.target.closest(".cell");
  if (!cell) return;
  const host = cell.parentElement.id === "pool"
    ? lastPool.filter(matchesColor)
    : visibleOwned(lastCards);
  placePreview(cell, host[+cell.dataset.i]);
});

document.addEventListener("mousemove", (e) => {
  if (e.target.closest(".cell") || e.target.closest(".drow") || e.target.closest("#preview")) return;
  preview.style.display = "none";
});

preview.addEventListener("click", (e) => {
  if (e.target.id !== "flip") return;
  const img = document.getElementById("pv");
  const showingBack = img.src.endsWith(img.dataset.back) || img.getAttribute("src") === img.dataset.back;
  img.src = showingBack ? img.dataset.front : img.dataset.back;
});

document.getElementById("q").addEventListener("input", () => loadCards().catch(err => say(err.message)));
document.getElementById("lookup").addEventListener("click", lookup);
document.getElementById("q").addEventListener("keydown", (e) => { if (e.key === "Enter") lookup(); });
document.getElementById("dq").addEventListener("input", () => refreshDeck().catch(err => say(err.message)));
document.getElementById("back").addEventListener("click", () => { show("decks"); loadDecks().catch(err => say(err.message)); });
document.getElementById("export").addEventListener("click", () => { window.location = "/api/export"; });
document.getElementById("import").addEventListener("change", async (e) => {
  const file = e.target.files[0];
  e.target.value = "";
  if (!file) return;
  const body = new FormData();
  body.append("file", file);
  if (!confirm("Replace the collection on this computer with that archive?")) return;
  const res = await fetch("/api/import", { method: "POST", body });
  if (!res.ok) { say("Import failed"); return; }
  say("Collection replaced from the archive.");
  await loadCards();
  if (document.getElementById("editor").classList.contains("on")) {
    try { await refreshDeck(); } catch { show("decks"); await loadDecks(); }
  } else {
    await loadDecks();
  }
});

async function lookup() {
  const q = document.getElementById("q").value.trim();
  if (!q) return;
  try {
    const prints = await api("GET", "/api/lookup?q=" + encodeURIComponent(q));
    document.getElementById("results").innerHTML = prints.map((p, i) => `
      <div class="hit">
        <img alt="" src="${p.faces && p.faces[0] ? p.faces[0].image_url : ""}">
        <div><strong>${escapeHTML(p.name)}</strong><div class="quiet">${escapeHTML(p.set_name)} · #${escapeHTML(p.number)} · ${escapeHTML(p.rarity)}</div>
        <div class="price">nonfoil ${p.price_nonfoil ? "$" + p.price_nonfoil : "—"} · foil ${p.price_foil ? "$" + p.price_foil : "—"}</div></div>
        <div class="actions">
          <button class="brass" type="button" data-i="${i}" data-foil="0">Add nonfoil ×1</button>
          <button class="brass" type="button" data-i="${i}" data-foil="1">Add foil ×1</button>
        </div>
      </div>`).join("") || `<p class="quiet">No printing found.</p>`;
    document.getElementById("results").querySelectorAll("button").forEach(btn => {
      btn.addEventListener("click", async () => {
        const p = prints[+btn.dataset.i];
        await api("POST", "/api/cards", { set: p.set, number: p.number, foil: btn.dataset.foil === "1", qty: 1 });
        say("Added " + p.name);
        await loadCards();
      });
    });
  } catch (err) {
    say(err.message);
  }
}

document.querySelectorAll(".topbar nav button").forEach(b => b.addEventListener("click", () => {
  show(b.dataset.view);
  if (b.dataset.view === "collection") loadCards().catch(err => say(err.message));
  if (b.dataset.view === "decks") loadDecks().catch(err => say(err.message));
}));

function rarityButtons(host) {
  host.innerHTML = rarityNames.map(r => `<button type="button" data-rarity="${r}">${r[0].toUpperCase() + r.slice(1)}</button>`).join("");
  host.querySelectorAll("button").forEach(btn => btn.addEventListener("click", () => {
    const r = btn.dataset.rarity;
    if (pickedRarity.has(r)) pickedRarity.delete(r); else pickedRarity.add(r);
    btn.classList.toggle("on", pickedRarity.has(r));
    renderOwned(lastCards);
  }));
}

colorButtons(document.getElementById("colors"));
colorButtons(document.getElementById("dcolors"));
rarityButtons(document.getElementById("rarities"));
document.querySelectorAll("#price-sort button").forEach(btn => btn.addEventListener("click", () => {
  const next = btn.dataset.sort;
  priceSort = priceSort === next ? "" : next;
  document.querySelectorAll("#price-sort button").forEach(b => b.classList.toggle("on", b.dataset.sort === priceSort));
  renderOwned(lastCards);
}));
loadCards().catch(err => say(err.message));
