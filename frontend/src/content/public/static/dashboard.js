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
	if (l === "CRIT" || l === "CRITICAL" || l === "FATAL" || l === "EMERG")
		return "pill-level-critical";
	if (l === "ERROR" || l === "ERR") return "pill-level-error";
	if (l === "WARN" || l === "WARNING") return "pill-level-warn";
	return "pill-level-info";
}

// Demo-mode classifier. Reads the nginx severity tag first (the
// authoritative source), then HTTP status codes for access logs, and
// finally keyword fallbacks. Keyword-only matching was too eager — the
// word "error" sits inside many non-error log messages.
function inferLevel(message) {
	if (!message) return "INFO";

	const tagMatch = message.match(
		/\[(emerg|alert|crit|error|warn|notice|info|debug)\]/i,
	);
	if (tagMatch) {
		const t = tagMatch[1].toLowerCase();
		if (t === "emerg" || t === "alert" || t === "crit") return "CRITICAL";
		if (t === "error") return "ERROR";
		if (t === "warn") return "WARN";
		return "INFO";
	}

	const statusMatch = message.match(/"\s+(\d{3})\s+/);
	if (statusMatch) {
		const code = parseInt(statusMatch[1], 10);
		if (code >= 500) return "ERROR";
		if (code >= 400) return "WARN";
		return "INFO";
	}

	if (/\b(panic|fatal|emerg|critical)\b/i.test(message)) return "CRITICAL";
	if (/\berror\b/i.test(message)) return "ERROR";
	if (/\bwarn(ing)?\b/i.test(message)) return "WARN";
	return "INFO";
}

// Pattern-based summarizer for the Message column. Real-Ollama summaries
// will flow in via `log.ai_summary` once `ai-core` is wired up on the
// homelab; this is the demo-mode fallback so the column has something
// useful to read regardless.
function summarize(message) {
	if (!message) return "—";

	const nf = message.match(/open\(\)\s+"([^"]+)"\s+failed\s+\(2:/);
	if (nf) {
		return `Nginx tried to serve ${nf[1]} but the file doesn't exist. Usually a probe or a stale link — safe to ignore if it's from a scanner.`;
	}
	if (/open\(\)\s+"[^"]+"\s+failed\s+\(13:\s*Permission/i.test(message)) {
		return "File exists but nginx can't read it. Check ownership and SELinux/AppArmor permissions on the path.";
	}
	if (/SSL_do_handshake\(\)\s+failed/i.test(message)) {
		return "TLS handshake failed. Likely a client using an unsupported protocol/cipher, or a misconfigured cert. Verify the SSL config.";
	}
	if (/worker process \d+ exited/i.test(message)) {
		return "Nginx worker exited. Could be normal config reload or an unexpected crash — check the exit signal.";
	}
	if (/upstream timed out/i.test(message)) {
		return "Upstream service didn't respond in time. Check the backend's health and `proxy_read_timeout`.";
	}
	if (/connect\(\)\s+failed\s+\(111:\s*Connection refused/i.test(message)) {
		return "Backend refused the connection. The upstream process is likely down or not listening on the expected port.";
	}

	const probe = message.match(
		/"GET\s+(\/\.git|\/\.env|\/admin|\/wp-admin|\/phpmyadmin|\/_profiler|\/geoserver|\/actuator)/i,
	);
	if (probe) {
		return `Scanner probing for ${probe[1]}. Standard attack-tool target. The 444 close is appropriate; ignore unless the rate spikes.`;
	}

	const status = message.match(/"\s+(\d{3})\s+/);
	if (status) {
		const code = parseInt(status[1], 10);
		if (code >= 500)
			return `Server returned ${code}. Application-side error — look at upstream logs for the root cause.`;
		if (code === 404)
			return "404 — resource not found. Either a broken link or a probe for a path that doesn't exist.";
		if (code === 403)
			return "403 — server refused the request. Could be IP block, auth failure, or a deny rule.";
		if (code === 401) return "401 — auth required. Client tried a protected route without credentials.";
		if (code === 444)
			return "444 — nginx closed the connection without a response. Typically used to drop malicious traffic.";
		if (code >= 400)
			return `Client error ${code}. Bad request, missing resource, or auth issue depending on path.`;
		if (code === 304) return "304 — cached response served. No action needed.";
		if (code >= 300) return `Redirect (${code}). Standard browser-initiated navigation.`;
		return `Successful request (${code}). Normal traffic.`;
	}

	return message.length > 140 ? message.slice(0, 140) + "…" : message;
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
	if (filter === "error") return l === "ERROR" || l === "ERR";
	if (filter === "critical")
		return l === "CRIT" || l === "CRITICAL" || l === "FATAL" || l === "EMERG";
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
	const summary = log.ai_summary || summarize(log.original_message);
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
		<td class="col-msg">
			<span class="msg-summary">${escapeHTML(summary)}</span>
			<span class="msg-raw" hidden>${escapeHTML(log.original_message)}</span>
		</td>
		<td class="col-toggle">
			<button class="msg-toggle" type="button" aria-label="Show raw log">Raw</button>
		</td>
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
			if (
				lvl === "ERROR" ||
				lvl === "ERR" ||
				lvl === "CRIT" ||
				lvl === "CRITICAL" ||
				lvl === "FATAL" ||
				lvl === "EMERG"
			) errorsSeen++;
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

// ---------- Summary/raw toggle (event delegation) ----------

logRowsEl.addEventListener("click", (e) => {
	const btn = e.target.closest(".msg-toggle");
	if (!btn) return;
	const row = btn.closest("tr");
	const summary = row.querySelector(".msg-summary");
	const raw = row.querySelector(".msg-raw");
	if (!summary || !raw) return;
	const summaryVisible = !summary.hidden;
	summary.hidden = summaryVisible;
	raw.hidden = !summaryVisible;
	btn.textContent = summaryVisible ? "Summary" : "Raw";
	btn.setAttribute(
		"aria-label",
		summaryVisible ? "Show summary" : "Show raw log",
	);
});

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
