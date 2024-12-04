const container = document.querySelector(".container");
const loginLink = document.querySelector(".login-link");
const registerLink = document.querySelector(".register-link");
const btnPopUp = document.querySelector(".log-in-button");
const iconClose = document.querySelector(".icon-close");

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
