const tabs = document.querySelectorAll(".auth-tab");
const forms = {
	login: document.getElementById("login-form"),
	register: document.getElementById("register-form"),
};

function activateForm(name) {
	tabs.forEach((tab) => {
		const active = tab.dataset.form === name;
		tab.classList.toggle("is-active", active);
		tab.setAttribute("aria-selected", active ? "true" : "false");
	});
	Object.entries(forms).forEach(([key, form]) => {
		if (!form) return;
		form.classList.toggle("is-hidden", key !== name);
	});
}

tabs.forEach((tab) => {
	tab.addEventListener("click", () => activateForm(tab.dataset.form));
});

// Password visibility toggle: flips input type on the targeted field
// and updates the eye icon. Applied to both login and register passwords.
document.querySelectorAll(".visibility-toggle").forEach((btn) => {
	btn.addEventListener("click", () => {
		const input = document.getElementById(btn.dataset.target);
		if (!input) return;
		const visible = input.type === "text";
		input.type = visible ? "password" : "text";
		btn.classList.toggle("is-visible", !visible);
		btn.setAttribute(
			"aria-label",
			visible ? "Show password" : "Hide password",
		);
	});
});

// Live password rule checker for the register form. Rules mirror the
// server-side validatePassword() in frontend/src/server.go.
const registerPw = document.getElementById("register-password");
const rulesEl = document.getElementById("password-rules");
if (registerPw && rulesEl) {
	const checks = {
		length: (v) => v.length >= 8,
		upper: (v) => /[A-Z]/.test(v),
		lower: (v) => /[a-z]/.test(v),
		number: (v) => /\d/.test(v),
		special: (v) => /[!@#$%^&*(),.?":{}|<>]/.test(v),
	};

	const ruleEls = {};
	rulesEl.querySelectorAll(".rule").forEach((el) => {
		ruleEls[el.dataset.rule] = el;
	});

	function updateRules(value) {
		for (const [name, check] of Object.entries(checks)) {
			const el = ruleEls[name];
			if (!el) continue;
			const met = check(value);
			el.classList.toggle("is-met", met);
			el.querySelector(".rule-icon").textContent = met ? "✓" : "○";
		}
	}

	const serverError = document.getElementById("password-error-register");

	registerPw.addEventListener("focus", () => {
		rulesEl.hidden = false;
	});
	registerPw.addEventListener("blur", () => {
		if (registerPw.value === "") rulesEl.hidden = true;
	});
	registerPw.addEventListener("input", () => {
		updateRules(registerPw.value);
		if (serverError) serverError.innerHTML = "";
	});
}
