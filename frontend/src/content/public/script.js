// Select elements
let container = document.querySelector(".container");
let loginLink = document.querySelector(".login-link");
let registerLink = document.querySelector(".register-link");
let btnPopUp = document.querySelector(".log-in-button");
let iconClose = document.querySelector(".icon-close");

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

// Close popup and reset forms when clicking the close icon
iconClose.addEventListener("click", () => {
	// Close the popup
	container.classList.remove("active-popup");

	// Reset the forms inside the container
	resetFormFields(".form-box.login form"); // Reset login form
	resetFormFields(".form-box.register form"); // Reset register form

	// Clear any validation error messages
	clearValidationErrors("#password-error-login"); // Clear login form error
	clearValidationErrors("#password-error-register"); // Clear register form error

	// Reset the "Remember Me" checkbox if it's checked
	const rememberMeCheckbox = document.querySelector(
		'.remember-forgot input[type="checkbox"]'
	);
	if (rememberMeCheckbox) {
		rememberMeCheckbox.checked = false;
	}
});

// Function to reset form fields
function resetFormFields(formSelector) {
	const form = document.querySelector(formSelector);
	if (form) {
		form.reset(); // Reset the form fields
	}
}

// Function to clear validation error messages
function clearValidationErrors(errorSelector) {
	const errorElement = document.querySelector(errorSelector);
	if (errorElement) {
		errorElement.textContent = ""; // Clear the error message
	}
}

// Password validation function
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
const loginButton = loginForm ? loginForm.querySelector(".btn") : null;
const passwordErrorLogin = loginForm
	? loginForm.querySelector("#password-error")
	: null;

if (loginButton && passwordErrorLogin) {
	loginButton.addEventListener("click", (e) => {
		const password = loginForm.querySelector(
			"input[type='password']"
		).value;

		// Clear previous error messages
		passwordErrorLogin.textContent = "";

		// Validate the password
		const errors = validatePassword(password);

		if (errors.length > 0) {
			e.preventDefault(); // Prevent form submission
			// Display the validation error messages
			passwordErrorLogin.textContent =
				"Password must meet the following criteria:\n" +
				errors.join("\n");
		}
	});
}

// Attach validation to the register button
const registerForm = document.querySelector(".form-box.register form");
const registerButton = registerForm ? registerForm.querySelector(".btn") : null;
const passwordErrorRegister = registerForm
	? registerForm.querySelector("#password-error")
	: null;

if (registerButton && passwordErrorRegister) {
	registerButton.addEventListener("click", (e) => {
		const password = registerForm.querySelector(
			"input[type='password']"
		).value;

		// Clear previous error messages
		passwordErrorRegister.textContent = "";

		// Validate the password
		const errors = validatePassword(password);

		if (errors.length > 0) {
			e.preventDefault();
			passwordErrorRegister.textContent =
				"Password must meet the following criteria:\n" +
				errors.join("\n");
		}
	});
}
