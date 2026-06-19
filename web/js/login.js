document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const errorMsg = document.getElementById('error-msg');
    errorMsg.classList.add('hidden');

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        const res = await fetch(window.BASE_PATH + '/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await res.json();

        if (!res.ok) {
            errorMsg.textContent = data.error || 'Login failed';
            errorMsg.classList.remove('hidden');
            return;
        }

        window.location.href = window.BASE_PATH + '/totp';
    } catch (err) {
        errorMsg.textContent = 'Network error';
        errorMsg.classList.remove('hidden');
    }
});
