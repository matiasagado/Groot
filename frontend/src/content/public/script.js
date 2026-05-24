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
