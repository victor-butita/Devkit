document.addEventListener('DOMContentLoaded', () => {
    // --- State & Navigation ---
    const toolContainers = document.querySelectorAll('.tool-container');
    const navLinks = document.querySelectorAll('.nav-link');

    function switchTool(toolId) {
        toolContainers.forEach(c => c.classList.add('hidden'));
        document.getElementById(`tool-${toolId}`).classList.remove('hidden');
        navLinks.forEach(l => l.classList.remove('active'));
        document.querySelector(`.nav-link[data-tool="${toolId}"]`).classList.add('active');
    }

    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            switchTool(e.currentTarget.dataset.tool);
        });
    });

    // --- API Helper ---
    async function apiCall(endpoint, body) {
        const response = await fetch(`/api${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'An unexpected error occurred.');
        }
        return data;
    }

    // --- Result Rendering ---
    function renderResult(resultEl, content, isCode = true) {
        resultEl.innerHTML = '';
        if (isCode) {
            const pre = document.createElement('pre');
            const code = document.createElement('code');
            code.textContent = content;
            pre.appendChild(code);
            resultEl.appendChild(pre);
        } else {
            resultEl.innerHTML = content;
        }
        resultEl.classList.remove('hidden');
    }

    // --- Tool: Mockify ---
    const mockifyInput = document.getElementById('mockify-input');
    const mockifySubmit = document.getElementById('mockify-submit');
    const mockifyResult = document.getElementById('mockify-result');

    mockifySubmit.addEventListener('click', async () => {
        try {
            const data = await apiCall('/mock/create', JSON.parse(mockifyInput.value));
            mockifyResult.innerHTML = `
                <div class="result-url">
                    <input type="text" value="${data.url}" readonly>
                    <button class="copy-btn">Copy</button>
                </div>`;
            mockifyResult.querySelector('.copy-btn').addEventListener('click', e => {
                navigator.clipboard.writeText(data.url);
                e.currentTarget.textContent = 'Copied!';
            });
            mockifyResult.classList.remove('hidden');
        } catch (err) {
            renderResult(mockifyResult, `<span style="color:var(--red);">${err.message}</span>`, false);
        }
    });

    // --- Tool: RegexCraft ---
    const regexInput = document.getElementById('regex-input');
    const regexSubmit = document.getElementById('regex-submit');
    const regexResult = document.getElementById('regex-result');

    regexSubmit.addEventListener('click', async () => {
        try {
            const data = await apiCall('/regex/generate', { description: regexInput.value });
            regexResult.innerHTML = `
                <pre><code>${data.regex}</code></pre>
                <div class="explanation">${data.explanation}</div>`;
            regexResult.classList.remove('hidden');
        } catch (err) {
            renderResult(regexResult, `<span style="color:var(--red);">${err.message}</span>`, false);
        }
    });

    // --- Tool: ConfigSwitch ---
    const configInput = document.getElementById('config-input');
    const configFrom = document.getElementById('config-from');
    const configTo = document.getElementById('config-to');
    const configOutput = document.getElementById('config-output');
    const configSubmit = document.getElementById('config-submit');

    configSubmit.addEventListener('click', async () => {
        try {
            const data = await apiCall('/config/convert', {
                input: configInput.value,
                from: configFrom.value,
                to: configTo.value,
            });
            configOutput.value = data.output;
        } catch (err) {
            configOutput.value = `Error: ${err.message}`;
        }
    });

    // --- Tool: QueryGen ---
    const sqlSchema = document.getElementById('sql-schema');
    const sqlDescription = document.getElementById('sql-description');
    const sqlSubmit = document.getElementById('sql-submit');
    const sqlResult = document.getElementById('sql-result');

    sqlSubmit.addEventListener('click', async () => {
        try {
            const data = await apiCall('/sql/generate', {
                schema: sqlSchema.value,
                description: sqlDescription.value,
            });
            renderResult(sqlResult, data.query);
        } catch (err) {
            renderResult(sqlResult, err.message, false);
        }
    });

    // --- Tool: JSON Beautifier ---
    const jsonInput = document.getElementById('json-input');
    const jsonSubmit = document.getElementById('json-submit');
    const jsonResult = document.getElementById('json-result');

    jsonSubmit.addEventListener('click', async () => {
        try {
            const data = await apiCall('/json/format', JSON.parse(jsonInput.value));
            renderResult(jsonResult, data.formatted_json);
        } catch (err) {
            renderResult(jsonResult, `<span style="color:var(--red);">Invalid JSON: ${err.message}</span>`, false);
        }
    });

    // --- Initial Setup ---
    switchTool('mockify');
});