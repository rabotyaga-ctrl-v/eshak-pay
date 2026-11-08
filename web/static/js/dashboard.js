const userEmail = document.getElementById("user-email");
const accountBalance = document.getElementById("account-balance");
const headerUserEmail = document.getElementById("header-user-email");
const logoutButton = document.getElementById("logout-button");
const transferPageButton =
    document.getElementById("transfer-page-button");

transferPageButton.addEventListener("click", () => {
    window.location.href = "/transfer-page";
});
const transactionsPageButton =
    document.getElementById("transactions-page-button");

transactionsPageButton.addEventListener("click", () => {
    window.location.href = "/transactions-page";
});

async function loadUser() {
    try {
        const response = await fetch("/me");

        if (!response.ok) {
            window.location.href = "/login";
            return;
        }

        const data = await response.json();

        userEmail.textContent = data.email;
        headerUserEmail.textContent = data.email;

    } catch (error) {
        window.location.href = "/login";
    }
}

async function loadAccount() {
    try {
        const response = await fetch("/account");

        if (!response.ok) {
            return;
        }

        const data = await response.json();

        const accountId = document.getElementById("account-id");
        const copyAccountIdButton = document.getElementById("copy-account-id");

        accountId.textContent = data.id;

        copyAccountIdButton.addEventListener("click", async () => {
            try {
                await navigator.clipboard.writeText(data.id);
                copyAccountIdButton.textContent = "Скопировано ✓";

                setTimeout(() => {
                    copyAccountIdButton.textContent = "Копировать";
                }, 1500);
            } catch (error) {
                console.error("Failed to copy account ID:", error);
            }
        });

        const balance = data.balance_kopecks / 100;

        accountBalance.textContent = `${balance.toLocaleString("ru-RU")} ₽`;


    } catch (error) {
        console.error("Failed to load account:", error);
    }
}

loadUser();
loadAccount();
logoutButton.addEventListener("click", async () => {
    try {
        await fetch("/logout", {
            method: "POST"
        });

        window.location.href = "/login";

    } catch (error) {
        console.error("Logout failed:", error);
    }
});