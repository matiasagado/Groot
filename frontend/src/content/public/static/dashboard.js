// Dashboard state + render loop.
// WebSocket payload shape:
//   { dt, level, ai_classified, original_message, platform, uuid }
//
// Architecture: incoming rows go into a master list (newest-first). The
// visible table is a filtered, paginated view of that list. Filters and
// pagination only re-render the view; the source data stays intact.

const PAGE_SIZE = 25;

let allRows = []; // newest first
let currentPage = 1;
let projectFilter = "all";
let severityFilter = "all";
let demoMode = false;
let projects = [];

let liveOn = true;
let totalSeen = 0;
let errorsSeen = 0;
let aiClassifiedSeen = 0;

const logRowsEl = document.getElementById("log-rows");
const liveIndicator = document.getElementById("live-indicator");
const pageRangeEl = document.getElementById("page-range");
const prevBtn = document.getElementById("page-prev");
const nextBtn = document.getElementById("page-next");
const projectChipsEl = document.getElementById("project-chips");
const severityChipsEl = document.getElementById("severity-chips");

// ---------- Helpers ----------

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
	const parts = dt.split(" ");
	return parts.length === 2 ? parts[1] : dt;
}

function levelClass(level) {
	const l = (level || "").toUpperCase();
	if (l === "ERROR" || l === "ERR" || l === "CRIT" || l === "FATAL")
		return "pill-level-error";
	if (l === "WARN" || l === "WARNING") return "pill-level-warn";
	return "pill-level-info";
}

// Demo-mode classifier: when the real AI hasn't tagged a row, infer a
// plausible severity from message content so recruiters see a populated
// AI column instead of an empty one.
function inferLevel(message) {
	if (!message) return "INFO";
	const statusMatch = message.match(/"\s+(\d{3})\s+/);
	if (statusMatch) {
		const code = parseInt(statusMatch[1], 10);
		if (code >= 500) return "ERROR";
		if (code >= 400) return "WARN";
		return "INFO";
	}
	if (/\b(crit|fatal|panic|exception|emerg)\b/i.test(message)) return "ERROR";
	if (/\berror\b/i.test(message)) return "ERROR";
	if (/\bwarn(ing)?\b/i.test(message)) return "WARN";
	return "INFO";
}

function resolveLevel(log) {
	if (log.ai_classified) return log.ai_classified;
	if (log.level) return log.level;
	if (demoMode) return inferLevel(log.original_message);
	return "";
}

function isAIClassified(log) {
	if (log.ai_classified) return true;
	if (demoMode) return true;
	return false;
}

function hashKey(s) {
	let h = 0;
	for (let i = 0; i < s.length; i++) {
		h = ((h << 5) - h) + s.charCodeAt(i);
		h |= 0;
	}
	return Math.abs(h);
}

function projectForLog(log) {
	if (demoMode && projects.length > 0) {
		const key = log.uuid || log.original_message || log.dt || "";
		return projects[hashKey(key) % projects.length];
	}
	return log.platform || "—";
}

function renderProjectChips(list) {
	if (!projectChipsEl) return;
	projectChipsEl.querySelectorAll(".chip").forEach((c) => {
		if (c.dataset.project !== "all") c.remove();
	});
	list.forEach((name) => {
		const btn = document.createElement("button");
		btn.type = "button";
		btn.className = "chip";
		btn.dataset.project = name;
		btn.textContent = name;
		projectChipsEl.appendChild(btn);
	});
}

// ---------- Filters + pagination ----------

function levelMatchesFilter(level, filter) {
	const l = level.toUpperCase();
	if (filter === "all") return true;
	if (filter === "info") return l === "INFO";
	if (filter === "warn") return l === "WARN" || l === "WARNING";
	if (filter === "error")
		return l === "ERROR" || l === "ERR" || l === "CRIT" || l === "FATAL";
	return true;
}

function applyFilters(rows) {
	return rows.filter((log) => {
		if (projectFilter !== "all" && projectForLog(log) !== projectFilter)
			return false;
		if (!levelMatchesFilter(resolveLevel(log), severityFilter)) return false;
		return true;
	});
}

function buildRow(log) {
	const level = resolveLevel(log).toUpperCase();
	const aiClassified = isAIClassified(log);
	const row = document.createElement("tr");
	if (log._isNew) {
		row.className = "new-log-flash";
		log._isNew = false;
		setTimeout(() => row.classList.remove("new-log-flash"), 1200);
	}
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
	return row;
}

function render() {
	const filtered = applyFilters(allRows);
	const total = filtered.length;
	const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
	if (currentPage > totalPages) currentPage = totalPages;
	const start = (currentPage - 1) * PAGE_SIZE;
	const end = Math.min(start + PAGE_SIZE, total);
	const visible = filtered.slice(start, end);

	logRowsEl.innerHTML = "";
	if (total === 0) {
		const empty = document.createElement("tr");
		empty.className = "log-row-empty";
		empty.innerHTML = `<td colspan="5">${
			allRows.length === 0
				? "Waiting for logs from the pipeline…"
				: "No logs match the current filters."
		}</td>`;
		logRowsEl.appendChild(empty);
	} else {
		visible.forEach((log) => logRowsEl.appendChild(buildRow(log)));
	}

	pageRangeEl.textContent =
		total === 0 ? "0 of 0" : `${start + 1}–${end} of ${total}`;
	prevBtn.disabled = currentPage <= 1;
	nextBtn.disabled = currentPage >= totalPages;
}

function setStat(id, value) {
	const el = document.getElementById(id);
	if (el) el.textContent = value;
}

function updateStats() {
	setStat("stat-total", totalSeen.toLocaleString());
	setStat("stat-errors", errorsSeen.toLocaleString());
	setStat("stat-recent", "—");
	const pct =
		totalSeen > 0
			? Math.round((aiClassifiedSeen / totalSeen) * 100) + "%"
			: "—";
	setStat("stat-ai", pct);
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
			const log = JSON.parse(event.data);
			log._isNew = true;
			allRows.unshift(log);
			totalSeen++;
			const lvl = resolveLevel(log).toUpperCase();
			if (lvl === "ERROR" || lvl === "CRIT" || lvl === "FATAL") errorsSeen++;
			if (isAIClassified(log)) aiClassifiedSeen++;
			updateStats();
			if (currentPage === 1) {
				render();
			} else {
				// keep pagination counts fresh without forcing a full table redraw
				const total = applyFilters(allRows).length;
				const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
				pageRangeEl.textContent = `${(currentPage - 1) * PAGE_SIZE + 1}–${Math.min(
					currentPage * PAGE_SIZE,
					total,
				)} of ${total}`;
				nextBtn.disabled = currentPage >= totalPages;
			}
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

// ---------- Chip handlers ----------

function setActiveChip(rowEl, chip) {
	rowEl.querySelectorAll(".chip").forEach((c) =>
		c.classList.toggle("is-active", c === chip),
	);
}

if (projectChipsEl) {
	projectChipsEl.addEventListener("click", (e) => {
		const chip = e.target.closest(".chip");
		if (!chip) return;
		projectFilter = chip.dataset.project;
		setActiveChip(projectChipsEl, chip);
		currentPage = 1;
		render();
	});
}

if (severityChipsEl) {
	severityChipsEl.addEventListener("click", (e) => {
		const chip = e.target.closest(".chip");
		if (!chip) return;
		severityFilter = chip.dataset.level;
		setActiveChip(severityChipsEl, chip);
		currentPage = 1;
		render();
	});
}

// ---------- Pagination ----------

prevBtn.addEventListener("click", () => {
	if (currentPage > 1) {
		currentPage--;
		render();
	}
});

nextBtn.addEventListener("click", () => {
	currentPage++;
	render();
});

// ---------- Search ⌘K focus ----------

document.addEventListener("keydown", (e) => {
	if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
		e.preventDefault();
		document.getElementById("search")?.focus();
	}
});

// ---------- Init ----------

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
		render();
	})
	.catch(() => render());

updateStats();
connectWebSocket();
