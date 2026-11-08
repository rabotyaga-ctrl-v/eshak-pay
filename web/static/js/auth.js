const registerForm = document.getElementById("register-form");

registerForm.addEventListener("submit", async (event) => {
    event.preventDefault();

    const email = document.getElementById("email").value;
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
        const response = await fetch("/register", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                email: email,
                password: password
            })
        });

        const responseText = await response.text();

        if (response.status === 201) {
            message.classList.add("success");
            message.textContent = "Аккаунт успешно создан. Теперь можно войти.";
        } else {
            message.classList.add("error");
            message.textContent = responseText;
        }

    } catch (error) {
        message.classList.add("error");
        message.textContent = "Не удалось связаться с сервером.";
    }
});