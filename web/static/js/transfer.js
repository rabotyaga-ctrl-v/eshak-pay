const transferForm = document.getElementById("transfer-form-element");

transferForm.addEventListener("submit", async (event) => {
    event.preventDefault();

    const recipientAccountId =
        document.getElementById("recipient-account-id").value.trim();

    const amount =
        document.getElementById("transfer-amount").value;

    const message = document.getElementById("transfer-message");

    message.className = "form-message";
    message.textContent = "";

    if (!recipientAccountId) {
        message.classList.add("error");
        message.textContent = "Введите счёт получателя.";
        return;
    }

    if (!amount || Number(amount) <= 0) {
        message.classList.add("error");
        message.textContent = "Введите сумму больше нуля.";
        return;
    }

    const amountKopecks = Math.round(Number(amount) * 100);

    const idempotencyKey = crypto.randomUUID();

    const response = await fetch("/transfer", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": idempotencyKey
        },
        body: JSON.stringify({
            recipient_account_id: recipientAccountId,
            amount_kopecks: amountKopecks
        })
    });

    const responseText = await response.text();

    let data = null;

    try {
        data = JSON.parse(responseText);
    } catch (error) {
        console.error("Transfer failed:", error);

        message.className = "form-message error";
        message.textContent = "Не удалось связаться с сервером.";
    }

    if (response.ok) {
        message.classList.add("success");
        message.textContent = "Перевод выполнен.";

        transferForm.reset();
        return;
    }

    message.classList.add("error");

    switch (response.status) {
        case 400:
            message.textContent = "Некорректные данные перевода.";
            break;

        case 409:
            message.textContent = "Конфликт идемпотентности.";
            break;

        case 422:
            message.textContent =
                data?.rejection_reason || "Перевод отклонён.";
            break;

        default:
            message.textContent = "Не удалось выполнить перевод.";
    }
});