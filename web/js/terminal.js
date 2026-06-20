const term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Monaco, Consolas, "Courier New", monospace',
    theme: {
        background: '#000000',
        foreground: '#ffffff'
    }
});

// Local copy/paste via Ctrl+Shift+C / Ctrl+Shift+V
term.attachCustomKeyEventHandler((e) => {
    // Mac: Cmd+Shift+C/V, Windows/Linux: Ctrl+Shift+C/V
    const isMod = navigator.platform.includes('Mac') ? e.metaKey : e.ctrlKey;
    if (isMod && e.shiftKey && e.type === 'keydown') {
        if (e.key === 'C' || e.key === 'c') {
            e.preventDefault();
            const selection = term.getSelection();
            if (selection) {
                navigator.clipboard.writeText(selection).catch(() => {});
                term.clearSelection();
            }
            return false;
        }
        if (e.key === 'V' || e.key === 'v') {
            e.preventDefault();
            navigator.clipboard.readText().then((text) => {
                if (text && ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({ type: 'data', payload: text }));
                }
            }).catch(() => {});
            return false;
        }
    }
    return true;
});

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));
fitAddon.fit();

// Right-click: 有选中则复制，无选中则粘贴
term.element.addEventListener('contextmenu', (e) => {
    e.preventDefault();
    const selection = term.getSelection();
    if (selection) {
        navigator.clipboard.writeText(selection).catch(() => {});
        term.clearSelection();
    } else {
        navigator.clipboard.readText().then((text) => {
            if (text && ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'data', payload: text }));
            }
        }).catch(() => {});
    }
});

let ws = null;
let reconnectAttempts = 0;
const MAX_RECONNECT = 3;

function connect() {
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${protocol}//${location.host}${window.BASE_PATH}/ws`);

    ws.onopen = () => {
        reconnectAttempts = 0;
        document.getElementById('reconnect-overlay').classList.add('hidden');
        term.focus();
    };

    ws.onmessage = (e) => {
        try {
            const msg = JSON.parse(e.data);
            if (msg.type === 'data') {
                term.write(msg.payload);
            } else if (msg.type === 'error') {
                term.write(`\x1b[31m${msg.payload}\x1b[0m`);
            } else if (msg.type === 'close') {
                ws.close();
            }
        } catch (err) {
            term.write(e.data);
        }
    };

    ws.onclose = () => {
        if (reconnectAttempts < MAX_RECONNECT) {
            reconnectAttempts++;
            setTimeout(connect, 1000 * reconnectAttempts);
        } else {
            document.getElementById('reconnect-overlay').classList.remove('hidden');
        }
    };

    ws.onerror = () => {
        ws.close();
    };
}

term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'data', payload: data }));
    }
});

window.addEventListener('resize', () => {
    fitAddon.fit();
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            type: 'resize',
            payload: { cols: term.cols, rows: term.rows }
        }));
    }
});

document.getElementById('reconnect-btn').addEventListener('click', () => {
    reconnectAttempts = 0;
    connect();
});

connect();

setTimeout(() => {
    fitAddon.fit();
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            type: 'resize',
            payload: { cols: term.cols, rows: term.rows }
        }));
    }
}, 100);
