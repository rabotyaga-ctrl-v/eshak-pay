const guestActions = document.getElementById("guest-actions");
const userActions = document.getElementById("user-actions");
const headerUserEmail = document.getElementById("header-user-email");
const logoutButton = document.getElementById("logout-button");

async function checkAuth() {
    try {
        const response = await fetch("/me");

        if (!response.ok) {
            guestActions.style.display = "flex";
            userActions.style.display = "none";
            return;
        }

        const data = await response.json();

        guestActions.style.display = "flex";
        userActions.style.display = "none";

        headerUserEmail.textContent = data.email;

        guestActions.style.display = "none";
        userActions.style.display = "flex";

    } catch (error) {
        guestActions.style.display = "flex";
        userActions.style.display = "none";
    }
}

logoutButton.addEventListener("click", async () => {
    try {
        await fetch("/logout", {
            method: "POST"
        });

        window.location.reload();

    } catch (error) {
        console.error("Logout failed:", error);
    }
});

checkAuth();