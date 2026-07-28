import http from 'node:http';
import { chromium } from 'playwright';

const listen = process.env.DISCOVERY_BROWSER_SIDECAR_LISTEN || process.env.ENGAGE_BROWSER_LISTEN || ':8910';
const port = Number.parseInt(listen.split(':').pop() || '8910', 10);

const TECH_PATTERNS = [
  { name: 'wordpress', re: /wp-content|wordpress/i },
  { name: 'react', re: /react|__NEXT_DATA__|_next\//i },
  { name: 'angular', re: /ng-version|angular/i },
  { name: 'jquery', re: /jquery/i },
  { name: 'nginx', re: /nginx/i },
  { name: 'apache', re: /apache/i },
  { name: 'php', re: /\.php|x-powered-by:\s*php/i },
];

function detectTechnologies(html, headers) {
  const found = new Set();
  const blob = html + ' ' + JSON.stringify(headers || {});
  for (const { name, re } of TECH_PATTERNS) {
    if (re.test(blob)) found.add(name);
  }
  const gen = html.match(/<meta[^>]+name=["']generator["'][^>]+content=["']([^"']+)["']/i);
  if (gen) found.add(gen[1].slice(0, 40));
  return [...found];
}

function analyzeSecurity(headers, pageUrl) {
  const issues = [];
  const h = {};
  for (const [k, v] of Object.entries(headers || {})) {
    h[k.toLowerCase()] = v;
  }
  const required = {
    'content-security-policy': 'CSP header missing (XSS mitigation)',
    'x-frame-options': 'X-Frame-Options missing (clickjacking risk)',
    'x-content-type-options': 'X-Content-Type-Options missing',
    'strict-transport-security': 'HSTS missing (HTTPS downgrade risk)',
  };
  for (const [key, desc] of Object.entries(required)) {
    if (!h[key]) {
      issues.push({ type: 'missing_security_header', severity: 'medium', description: desc, header: key });
    }
  }
  const csp = h['content-security-policy'] || '';
  if (csp && csp.includes('unsafe-inline')) {
    issues.push({ type: 'weak_csp', severity: 'low', description: 'CSP allows unsafe-inline scripts' });
  }
  if (pageUrl.startsWith('https://') && h['strict-transport-security'] === undefined) {
    // already reported
  }
  const score = Math.max(0, 100 - issues.length * 8);
  return {
    issues,
    total_issues: issues.length,
    security_score: score,
    modules: issues.length ? ['security_headers'] : [],
  };
}

async function runActiveTests(page, url) {
  const issues = [];
  try {
    const probe = `${url}${url.includes('?') ? '&' : '?'}engage_xss_probe=<script>engage</script>`;
    const resp = await page.goto(probe, { waitUntil: 'domcontentloaded', timeout: 15_000 });
    const body = await page.content();
    if (body.includes('<script>engage</script>') && !body.includes('&lt;script&gt;')) {
      issues.push({
        type: 'reflected_xss_candidate',
        severity: 'high',
        description: 'Probe string reflected without encoding (lab verification only)',
      });
    }
    if (resp) await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 15_000 });
  } catch {
    // ignore active test failures
  }
  return issues;
}

async function inspectPage(opts) {
  const {
    target,
    waitTime = 5,
    headless = true,
    wantScreenshot = true,
    activeTests = false,
  } = opts;
  const browser = await chromium.launch({ headless });
  const page = await browser.newPage();
  let resp;
  try {
    resp = await page.goto(target, { waitUntil: 'domcontentloaded', timeout: 60_000 });
  } catch (err) {
    await browser.close();
    return { success: false, error: String(err) };
  }
  if (waitTime > 0) {
    await page.waitForTimeout(Math.min(waitTime, 30) * 1000);
  }

  const title = await page.title();
  const url = page.url();
  const status = resp?.status() ?? 0;
  const headers = resp?.headers() ?? {};

  const forms = await page.$$eval('form', (nodes) =>
    nodes.slice(0, 20).map((f) => ({
      action: f.getAttribute('action') || '',
      method: (f.getAttribute('method') || 'get').toLowerCase(),
      inputs: [...f.querySelectorAll('input,textarea,select')].slice(0, 15).map((i) => ({
        name: i.getAttribute('name') || '',
        type: i.getAttribute('type') || i.tagName.toLowerCase(),
      })),
    }))
  );
  const links = await page.$$eval('a[href]', (nodes) =>
    nodes.slice(0, 50).map((a) => ({ href: a.getAttribute('href') || '', text: (a.textContent || '').trim().slice(0, 80) }))
  );
  const inputs = await page.$$eval('input,textarea,select', (nodes) =>
    nodes.slice(0, 30).map((i) => ({
      name: i.getAttribute('name') || '',
      type: i.getAttribute('type') || i.tagName.toLowerCase(),
    }))
  );
  const scripts = await page.$$eval('script[src]', (nodes) =>
    nodes.slice(0, 20).map((s) => s.getAttribute('src') || '')
  );

  const html = await page.content();
  let security = analyzeSecurity(headers, url);
  if (activeTests) {
    const active = await runActiveTests(page, url);
    security.issues.push(...active);
    security.total_issues = security.issues.length;
    security.security_score = Math.max(0, 100 - security.total_issues * 8);
    if (active.length) security.modules.push('active_tests');
  }

  const technologies = detectTechnologies(html, headers);
  let screenshot = '';
  if (wantScreenshot) {
    screenshot = (await page.screenshot({ type: 'png' })).toString('base64');
  }
  await browser.close();

  const securitySubset = {
    security_score: security.security_score,
    total_issues: security.total_issues,
    modules: security.modules.slice(0, 10),
    issues: security.issues.slice(0, 25),
    issues_truncated: security.issues.length > 25,
  };

  const result = {
    success: true,
    forms,
    page_info: { title, url, status, forms, links, inputs, scripts },
    security_analysis: securitySubset,
    technologies,
    screenshot,
    timestamp: new Date().toISOString(),
  };
  return result;
}

async function execBrowser(target, wantScreenshot) {
  const r = await inspectPage({ target, wantScreenshot, waitTime: 2 });
  if (!r.success) {
    return { stdout: '', stderr: r.error || '', exit_code: 1, error: r.error || 'inspect failed' };
  }
  return {
    stdout: `${JSON.stringify(r)}\n`,
    stderr: '',
    exit_code: 0,
    error: '',
  };
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', (c) => chunks.push(c));
    req.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
    req.on('error', reject);
  });
}

const server = http.createServer(async (req, res) => {
  try {
    if (req.url === '/health' && req.method === 'GET') {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ ok: true, service: 'discovery-browser-agent', engine: 'playwright' }));
      return;
    }
    if ((req.url === '/inspect' || req.url === '/exec') && req.method === 'POST') {
      const raw = await readBody(req);
      const data = raw ? JSON.parse(raw) : {};
      const target = (data.url || data.target || (data.args && data.args[0]) || '').trim();
      if (!target) {
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ success: false, error: 'target required' }));
        return;
      }
      if (req.url === '/inspect' || data.inspect) {
        const waitTime = Number.parseInt(data.wait_time ?? data.waitTime ?? '5', 10) || 5;
        const headless = data.headless !== false && data.headless !== 'false';
        const result = await inspectPage({
          target,
          waitTime,
          headless,
          wantScreenshot: data.screenshot !== false,
          activeTests: Boolean(data.active_tests || data.activeTests),
        });
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(result));
        return;
      }
      const out = await execBrowser(target, Boolean(data.screenshot));
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(out));
      return;
    }
    res.writeHead(404);
    res.end('not found');
  } catch (err) {
    res.writeHead(500, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ success: false, error: String(err) }));
  }
});

server.listen(port, '0.0.0.0', () => {
  console.log(`discovery browser-agent (playwright) on :${port}`);
});
