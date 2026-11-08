const transactionsList =
    document.getElementById("transactions-list");
const limitButtons =
    document.querySelectorAll(".transactions-controls button");
const transactionsMessage =
    document.getElementById("transactions-message");

async function loadTransactions(limit = 20) {
    try {
        const response = await fetch(`/transactions?limit=${limit}`);

        if (!response.ok) {
            if (response.status === 401) {
                window.location.href = "/login";
                return;
            }

            throw new Error(`HTTP ${response.status}`);
        }

        const data = await response.json();

        console.log(data);

        transactionsList.innerHTML = "";

        transactionsList.innerHTML = "";

        transactionsList.innerHTML = "";

        const moneyFormatter = new Intl.NumberFormat("ru-RU", {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2
        });

        data.transactions.forEach(transaction => {
            const item = document.createElement("div");

            item.className = "transaction-item";

            const amount = transaction.amount_kopecks / 100;

            const title =
                transaction.type === "TRANSFER"
                    ? "Перевод"
                    : "Пополнение";

            const sign = amount > 0 ? "+" : "−";

            const amountClass =
            amount > 0
                ? "transaction-income"
                : "transaction-expense";

            const formattedAmount =
            `${sign}${moneyFormatter.format(Math.abs(amount))} ₽`;
            
            const date = new Date(transaction.created_at);

            const formattedDate = date.toLocaleString("ru-RU", {
                day: "2-digit",
                month: "long",
                hour: "2-digit",
                minute: "2-digit"
            });

            item.innerHTML = `
                <div class="transaction-info">
                    <span class="transaction-title">${title}</span>
                    <span class="transaction-date">${formattedDate}</span>
                </div>

                <span class="transaction-amount ${amountClass}">
                    ${formattedAmount}
                </span>
            `;

            transactionsList.appendChild(item);
        });
    } catch (error) {
        console.error("Failed to load transactions:", error);

        transactionsList.innerHTML = "";

        transactionsMessage.className = "form-message error";
        transactionsMessage.textContent =
            "Не удалось загрузить историю операций.";
    }
}

loadTransactions();

limitButtons[0].classList.add("active");

limitButtons.forEach(button => {
    button.addEventListener("click", () => {
        const limit = Number(button.dataset.limit);

        limitButtons.forEach(button => {
            button.classList.remove("active");
        });

        button.classList.add("active");

        loadTransactions(limit);
    });
});