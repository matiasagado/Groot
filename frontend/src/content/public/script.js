// DOM elements
const container = document.querySelector(".container");
const loginLink = document.querySelector(".login-link");
const registerLink = document.querySelector(".register-link");
const btnPopUp = document.querySelector(".log-in-button");
const iconClose = document.querySelector(".icon-close");

// Event listeners for toggling between login and register forms
registerLink.addEventListener("click", () => {
	container.classList.add("active"); // Switch to register form
});

loginLink.addEventListener("click", () => {
	container.classList.remove("active"); // Switch to login form
});

btnPopUp.addEventListener("click", () => {
	container.classList.add("active-popup"); // Show popup
});

iconClose.addEventListener("click", () => {
	container.classList.remove("active-popup"); // Close popup
});

// Password validation logic
const validatePassword = (password) => {
	const criteria = [
		{ regex: /.{8,}/, message: "At least 8 characters" },
		{ regex: /[A-Z]/, message: "At least one uppercase letter" },
		{ regex: /[a-z]/, message: "At least one lowercase letter" },
		{ regex: /\d/, message: "At least one number" },
		{
			regex: /[!@#$%^&*(),.?":{}|<>]/,
			message: "At least one special character",
		},
	];

	// Return an array of unmet criteria
	return criteria
		.filter((criterion) => !criterion.regex.test(password))
		.map((criterion) => criterion.message);
};

// Attach validation to the login form
const loginForm = document.querySelector("#login-form");
const passwordErrorLogin = loginForm.querySelector("#password-error-login");

loginForm.addEventListener("submit", (e) => {
	const password = loginForm.querySelector("#password").value;

	// Clear previous error messages
	passwordErrorLogin.textContent = "";

	// Validate the password
	const errors = validatePassword(password);

	if (errors.length > 0) {
		e.preventDefault(); // Prevent form submission
		passwordErrorLogin.innerHTML = errors
			.map((error) => `<p>${error}</p>`)
			.join(""); // Display errors as separate lines
	}
});

// Attach validation to the register form
const registerForm = document.querySelector("#register-form");
const passwordErrorRegister = registerForm.querySelector(
	"#password-error-register"
);

registerForm.addEventListener("submit", (e) => {
	const password = registerForm.querySelector("#register-password").value;

	// Clear previous error messages
	passwordErrorRegister.textContent = "";

	// Validate the password
	const errors = validatePassword(password);

	if (errors.length > 0) {
		e.preventDefault(); // Prevent form submission
		passwordErrorRegister.innerHTML = errors
			.map((error) => `<p>${error}</p>`)
			.join(""); // Display errors as separate lines
	}
});


# a test