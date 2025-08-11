document.addEventListener('DOMContentLoaded', () => {
    // --- State & Navigation ---
    const mainContent = document.querySelector('.main-content');
    const navLinks = document.querySelectorAll('.nav-link');
    const pageTitle = document.getElementById('page-title');

    function switchTool(toolId) {
        mainContent.innerHTML = '';
        const template = document.getElementById(`template-${toolId}`);
        if (!template) { console.error(`Template not found: ${toolId}`); return; }
        const clone = template.content.cloneNode(true);
        mainContent.appendChild(clone);

        navLinks.forEach(l => l.classList.remove('active'));
        const activeLink = document.querySelector(`.nav-link[data-tool="${toolId}"]`);
        activeLink.classList.add('active');
        pageTitle.textContent = activeLink.textContent;

        initializeTool(toolId);
    }

    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            switchTool(e.currentTarget.dataset.tool);
        });
    });

    // --- API Helper ---
    async function apiCall(endpoint, body, submitButton) {
        const originalText = submitButton.textContent;
        submitButton.textContent = 'Processing...';
        submitButton.disabled = true;
        try {
            const response = await fetch(`/api${endpoint}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), });
            const data = await response.json();
            if (!response.ok) throw new Error(data.error || 'An unexpected error occurred.');
            return data;
        } finally {
            submitButton.textContent = originalText;
            submitButton.disabled = false;
        }
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
        } else { resultEl.innerHTML = content; }
        resultEl.classList.remove('hidden');
    }
    function renderError(resultEl, message) { renderResult(resultEl, `<span style="color:var(--red); font-weight:500;">${message}</span>`, false); }

    // --- Universal Tool Initializer ---
    function initializeTool(toolId) {
        const toolContent = mainContent.querySelector('.tool-content');
        if (!toolContent) return;
        switch (toolId) {
            case 'mockify': initMockify(toolContent); break;
            case 'regex': initRegex(toolContent); break;
            case 'config': initConfig(toolContent); break;
            case 'sql': initSql(toolContent); break;
            case 'json': initJson(toolContent); break;
        }
    }

    // --- Tool-specific initializers ---
    function initMockify(container) {
        const input = container.querySelector('#mockify-input');
        const submit = container.querySelector('#mockify-submit');
        const result = container.querySelector('#mockify-result');
        submit.addEventListener('click', async () => {
            try {
                const data = await apiCall('/mock/create', JSON.parse(input.value), submit);
                result.innerHTML = `<div class="result-url"><input type="text" value="${data.url}" readonly><button class="copy-btn">Copy</button></div>`;
                result.querySelector('.copy-btn').addEventListener('click', e => {
                    navigator.clipboard.writeText(data.url);
                    const btn = e.currentTarget;
                    btn.textContent = 'Copied!';
                    btn.classList.add('copied');
                    setTimeout(() => { btn.textContent = 'Copy'; btn.classList.remove('copied'); }, 2000);
                });
                result.classList.remove('hidden');
            } catch (err) { renderError(result, err.message); }
        });
    }

    function initRegex(container) {
        const input = container.querySelector('#regex-input');
        const submit = container.querySelector('#regex-submit');
        const result = container.querySelector('#regex-result');
        submit.addEventListener('click', async () => {
            try {
                const data = await apiCall('/regex/generate', { description: input.value }, submit);
                result.innerHTML = `<pre><code>${data.regex}</code></pre><div class="explanation">${data.explanation}</div>`;
                result.classList.remove('hidden');
            } catch (err) { renderError(result, err.message); }
        });
    }

    function initConfig(container) {
        const input = container.querySelector('#config-input');
        const fromGroup = container.querySelector('#config-from');
        const toGroup = container.querySelector('#config-to');
        const output = container.querySelector('#config-output');
        const submit = container.querySelector('#config-submit');

        // Logic for custom button groups
        fromGroup.addEventListener('click', e => {
            if (e.target.tagName === 'BUTTON') {
                fromGroup.querySelector('.active').classList.remove('active');
                e.target.classList.add('active');
            }
        });
        toGroup.addEventListener('click', e => {
            if (e.target.tagName === 'BUTTON') {
                toGroup.querySelector('.active').classList.remove('active');
                e.target.classList.add('active');
            }
        });

        submit.addEventListener('click', async () => {
            try {
                const fromVal = fromGroup.querySelector('.active').dataset.value;
                const toVal = toGroup.querySelector('.active').dataset.value;
                const data = await apiCall('/config/convert', { input: input.value, from: fromVal, to: toVal }, submit);
                output.value = data.output;
            } catch (err) { output.value = `Error: ${err.message}`; }
        });
    }

    function initSql(container) {
        const schema = container.querySelector('#sql-schema');
        const description = container.querySelector('#sql-description');
        const submit = container.querySelector('#sql-submit');
        const result = container.querySelector('#sql-result');
        submit.addEventListener('click', async () => {
            try {
                const data = await apiCall('/sql/generate', { schema: schema.value, description: description.value }, submit);
                renderResult(result, data.query);
            } catch (err) { renderError(result, err.message); }
        });
    }

    function initJson(container) {
        const input = container.querySelector('#json-input');
        const submit = container.querySelector('#json-submit');
        const result = container.querySelector('#json-result');
        submit.addEventListener('click', async () => {
            try {
                const data = await apiCall('/json/format', JSON.parse(input.value), submit);
                renderResult(result, data.formatted_json);
            } catch (err) { renderError(result, `Invalid JSON: ${err.message}`); }
        });
    }

    // --- Initial Load ---
    switchTool('mockify');
});