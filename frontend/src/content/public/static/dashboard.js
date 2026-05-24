// Dashboard live-log feed.
// WebSocket payload shape: { dt, level, original_message, platform }.
// Project chips / severity chips / search input are scaffolded — they look
// real and respond to clicks, but don't filter until the backend supports
// project-labeled data and a filter endpoint.

const MAX_VISIBLE_ROWS = 5000;
const logRowsEl = document.getElementById("log-rows");
const liveIndicator = document.getElementById("live-indicator");

let liveOn = true;
let totalSeen = 0;
let errorsSeen = 0;
let aiClassifiedSeen = 0;

function levelClass(level) {
	const l = (level || "").toUpperCase();
	if (l === "ERROR" || l === "ERR") return "pill-level-error";
	if (l === "WARN" || l === "WARNING") return "pill-level-warn";
	return "pill-level-info";
}

function escapeHTML(s) {
	if (s == null) return "";
	return String(s)
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#39;");
}

function formatTime(dt) {
	if (!dt) return "—";
	// "2026-05-23 14:23:01" → "14:23:01" with date subtle on hover
	const parts = dt.split(" ");
	return parts.length === 2 ? parts[1] : dt;
}

// Demo mode + projects come from /api/me; defaults are safe for the
// pre-fetch render path.
let demoMode = false;
let projects = [];

function hashKey(s) {
	let h = 0;
	for (let i = 0; i < s.length; i++) {
		h = ((h << 5) - h) + s.charCodeAt(i);
		h |= 0;
	}
	return Math.abs(h);
}

function projectForLog(log) {
	// In demo mode, deterministically assign each row to one of the mock
	// projects so the chips/rows look like they came from a real fleet.
	if (demoMode && projects.length > 0) {
		const key = log.original_message || log.dt || "";
		return projects[hashKey(key) % projects.length];
	}
	return log.platform || "—";
}

function renderProjectChips(list) {
	const row = document.getElementById("project-chips");
	if (!row) return;
	row.querySelectorAll(".chip").forEach((c) => {
		if (c.dataset.project !== "all") c.remove();
	});
	list.forEach((name) => {
		const btn = document.createElement("button");
		btn.type = "button";
		btn.className = "chip";
		btn.dataset.project = name;
		btn.textContent = name;
		row.appendChild(btn);
	});
}

function setStat(id, value) {
	const el = document.getElementById(id);
	if (el) el.textContent = value;
}

function updateStats() {
	setStat("stat-total", totalSeen.toLocaleString());
	setStat("stat-errors", errorsSeen.toLocaleString());
	setStat("stat-recent", "—"); // needs server-side aggregation
	const pct =
		totalSeen > 0
			? Math.round((aiClassifiedSeen / totalSeen) * 100) + "%"
			: "—";
	setStat("stat-ai", pct);
}

function clearEmptyRow() {
	const empty = logRowsEl.querySelector(".log-row-empty");
	if (empty) empty.remove();
}

function appendLogRow(log) {
	clearEmptyRow();

	const level = (log.level || "").toUpperCase();
	const aiClassified = level === "INFO" || level === "ERROR";

	const row = document.createElement("tr");
	row.className = "new-log-flash";
	row.innerHTML = `
		<td class="col-time" title="${escapeHTML(log.dt)}">${escapeHTML(
		formatTime(log.dt),
	)}</td>
		<td class="col-project">
			<span class="pill">${escapeHTML(projectForLog(log))}</span>
		</td>
		<td class="col-level">
			${
				level
					? `<span class="pill ${levelClass(level)}">${escapeHTML(
							level,
					  )}</span>`
					: `<span class="pill pill-ai-no">—</span>`
			}
		</td>
		<td class="col-ai">
			<span class="pill ${aiClassified ? "pill-ai-yes" : "pill-ai-no"}">
				${aiClassified ? "yes" : "no"}
			</span>
		</td>
		<td class="col-msg">${escapeHTML(log.original_message)}</td>
	`;
	logRowsEl.insertAdjacentElement("afterbegin", row);
	setTimeout(() => row.classList.remove("new-log-flash"), 1200);

	totalSeen++;
	if (level === "ERROR") errorsSeen++;
	if (aiClassified) aiClassifiedSeen++;
	updateStats();

	const rows = logRowsEl.querySelectorAll("tr");
	if (rows.length > MAX_VISIBLE_ROWS) rows[rows.length - 1].remove();
}

// ---------- WebSocket feed ----------

function connectWebSocket() {
	const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
	const ws = new WebSocket(
		proto + "//" + window.location.host + "/classifiedLogsWebSocket",
	);
	ws.onopen = () => console.log("WebSocket connected");
	ws.onmessage = (event) => {
		if (!liveOn) return;
		try {
			appendLogRow(JSON.parse(event.data));
		} catch (e) {
			console.error("Failed to parse log payload:", e);
		}
	};
	ws.onerror = (err) => console.error("WebSocket error:", err);
	ws.onclose = () => {
		console.log("WebSocket closed, retrying in 3s");
		setTimeout(connectWebSocket, 3000);
	};
}

// ---------- Live toggle ----------

if (liveIndicator) {
	liveIndicator.addEventListener("click", () => {
		liveOn = !liveOn;
		liveIndicator.classList.toggle("is-active", liveOn);
		liveIndicator.setAttribute("aria-pressed", liveOn ? "true" : "false");
		liveIndicator.querySelector(".live-label").textContent = liveOn
			? "Live"
			: "Paused";
	});
}

// ---------- Chip toggles (visual only) ----------

document.querySelectorAll(".chip-row").forEach((row) => {
	row.addEventListener("click", (e) => {
		const chip = e.target.closest(".chip");
		if (!chip) return;
		row.querySelectorAll(".chip").forEach((c) =>
			c.classList.toggle("is-active", c === chip),
		);
	});
});

// ---------- Search ⌘K focus ----------

document.addEventListener("keydown", (e) => {
	if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
		e.preventDefault();
		document.getElementById("search")?.focus();
	}
});

// Ask the server whether this session is in demo mode, what projects to
// show, and surface the banner / chips accordingly.
fetch("/api/me")
	.then((r) => r.json())
	.then((me) => {
		demoMode = !!me.demo;
		projects = Array.isArray(me.projects) ? me.projects : [];
		if (demoMode) {
			document.body.classList.add("is-demo");
			const banner = document.getElementById("demo-banner");
			if (banner) banner.hidden = false;
		}
		renderProjectChips(projects);
	})
	.catch(() => {});

connectWebSocket();
updateStats();
