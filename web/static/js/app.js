
const themeToggle = document.getElementById("theme-toggle");

// Применяем сохранённую тему
const savedTheme = localStorage.getItem("theme") || "light";

document.documentElement.dataset.theme = savedTheme;

// Если кнопка переключения темы присутствует на странице
if (themeToggle) {
    themeToggle.textContent = savedTheme === "dark" ? "☀" : "◐";

    themeToggle.addEventListener("click", () => {
        const currentTheme = document.documentElement.dataset.theme;
        const newTheme = currentTheme === "light" ? "dark" : "light";

        document.documentElement.dataset.theme = newTheme;
        localStorage.setItem("theme", newTheme);

        themeToggle.textContent = newTheme === "dark" ? "☀" : "◐";
    });
}
