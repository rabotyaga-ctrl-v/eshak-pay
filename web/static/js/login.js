const loginForm = document.getElementById("login-form");

loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();

    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;
    const message = document.getElementById("form-message");

    message.className = "form-message";
    message.textContent = "";

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    if (!emailRegex.test(email)) {
        message.classList.add("error");
        message.textContent = "Введите корректный email.";
        return;
    }

    if (password.length < 8) {
        message.classList.add("error");
        message.textContent = "Пароль должен содержать минимум 8 символов.";
        return;
    }

    try {
        const response = await fetch("/login", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                email: email,
                password: password
            })
        });

        if (response.ok) {
            window.location.href = "/dashboard";
        } else {
            const responseText = await response.text();

            message.classList.add("error");

            if (response.status === 401) {
                message.textContent = "Неверный email или пароль.";
            } else {
                message.textContent = responseText || "Не удалось войти.";
            }
        }

    } catch (error) {
        message.classList.add("error");
        message.textContent = "Не удалось связаться с сервером.";
    }
});