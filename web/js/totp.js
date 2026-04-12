document.getElementById('totp-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const errorMsg = document.getElementById('error-msg');
    errorMsg.classList.add('hidden');

    const code = document.getElementById('totp-code').value;

    try {
        const res = await fetch('/api/totp', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code })
        });

        const data = await res.json();

        if (!res.ok) {
            errorMsg.textContent = data.error || 'Verification failed';
            errorMsg.classList.remove('hidden');
            return;
        }

        window.location.href = '/terminal';
    } catch (err) {
        errorMsg.textContent = 'Network error';
        errorMsg.classList.remove('hidden');
    }
});
