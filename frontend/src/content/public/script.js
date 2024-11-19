const container = document.querySelector(".container");
const loginLink = document.querySelector(".login-link");
const registerLink = document.querySelector(".register-link");
const btnPopUp = document.querySelector(".log-in-button");
const iconClose = document.querySelector(".icon-close");

// Event listeners for toggling between login and register forms
registerLink.addEventListener("click", () => {
	container.classList.add("active");
});

loginLink.addEventListener("click", () => {
	container.classList.remove("active");
});

btnPopUp.addEventListener("click", () => {
	container.classList.add("active-popup");
});

iconClose.addEventListener("click", () => {
	container.classList.remove("active-popup");
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

	const errors = criteria
		.filter((criterion) => !criterion.regex.test(password))
		.map((criterion) => criterion.message);

	return errors;
};

// Attach validation to the login button
const loginForm = document.querySelector(".form-box.login form");
const loginButton = loginForm.querySelector(".btn");
const passwordErrorLogin = loginForm.querySelector("#password-error");

loginButton.addEventListener("click", (e) => {
	const password = loginForm.querySelector("input[type='password']").value;

	// Clear previous error messages
	passwordErrorLogin.textContent = "";

	// Validate the password
	const errors = validatePassword(password);

	if (errors.length > 0) {
		e.preventDefault(); // Prevent form submission
		// Display the validation error messages
		passwordErrorLogin.textContent =
			"Password must meet the following criteria:\n" + errors.join("\n");
	}
});

// Attach validation to the register button
const registerForm = document.querySelector(".form-box.register form");
const registerButton = registerForm.querySelector(".btn");
const passwordErrorRegister = registerForm.querySelector("#password-error");

registerButton.addEventListener("click", (e) => {
	const password = registerForm.querySelector("input[type='password']").value;

	// Clear previous error messages
	passwordErrorRegister.textContent = "";

	// Validate the password
	const errors = validatePassword(password);

	if (errors.length > 0) {
		e.preventDefault(); // Prevent form submission
		// Display the validation error messages
		passwordErrorRegister.textContent =
			"Password must meet the following criteria:\n" + errors.join("\n");
	}
});
