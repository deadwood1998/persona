// Exercise selection, draft editing, and send recovery on the real chat with an intercepted hub.
// node web/e2e/inline-text-comments.mjs <loopback Vite origin> <artifacts>
import assert from "node:assert/strict";
import { mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import { chromium } from "playwright-core";

const [address, artifacts] = process.argv.slice(2);
assert.ok(address && artifacts);
const base = new URL(address);
const output = resolve(artifacts);
await mkdir(output, { recursive: true });
const AT = "2026-01-01T10:00:00Z";
const MACHINE = { id: "example", name: "Computer A", online: true, os: "linux" };
const SESSION = { machine_id: MACHINE.id, session_id: "example-main", title: "Persona", state: "idle", status: "idle", cwd: "/example", updated_at: AT, assistant: true, started_by_persona: true, capabilities: { send: true, interrupt: true } };
const ITEMS = [
  { id: "u1", seq: 1, role: "user", kind: "text", text: "Please review this example plan.", ts: AT },
  { id: "a1", seq: 2, role: "assistant", kind: "text", text: "The example plan has three steps: prepare, check and deliver.", ts: AT },
  { id: "u2", seq: 3, role: "user", kind: "text", text: "What happens after the first check?", ts: AT },
  { id: "a2", seq: 4, role: "assistant", kind: "text", text: "Messages after the quoted passage should remain visible while you write your comments.", ts: AT },
];
assert.ok(["127.0.0.1", "localhost"].includes(base.hostname));
const browser = await chromium.launch({ channel: "chrome" });
const deadline = setTimeout(() => void browser.close(), 120_000);
try {
  for (const width of [1440, 390]) for (const theme of ["light", "dark"]) {
    const context = await browser.newContext({ viewport: { width, height: 800 }, colorScheme: theme, reducedMotion: "reduce", serviceWorkers: "block" });
    await context.addInitScript((theme) => {
      localStorage.setItem("persona.token", "example");
      localStorage.setItem("persona.language", "en");
      localStorage.setItem("persona.theme", theme);
    }, theme);
    await context.route("**/*", (route) => {
      if (new URL(route.request().url()).origin !== base.origin) return route.abort();
      return route.continue();
    });
    const writes = [];
    let reject = true;
    await context.route("**/api/v1/**", (route) => {
      const method = route.request().method();
      if (method !== "GET") {
        if (new URL(route.request().url()).pathname.endsWith("/messages")) {
          writes.push(route.request().postDataJSON());
          if (reject) return route.fulfill({ status: 400, json: { error: "Example failure", delivered: "none" } });
        }
        return route.fulfill({ status: 200, json: {} });
      }
      const path = new URL(route.request().url()).pathname.replace("/api/v1", "");
      const data = path === "/machines" ? { machines: [MACHINE] }
        : path === "/sessions" ? { sessions: [SESSION] }
        : path === "/assistant" ? { machine_id: MACHINE.id, main: SESSION.session_id, conversations: [SESSION], tasks: [], memory: [] }
        : path === "/interactions" ? { interactions: [] }
        : path.endsWith("/items") ? { items: ITEMS, has_more: false }
        : path.endsWith("/feed") ? { items: [], plans: {}, task_cards: [] }
        : path === "/todos" ? { todos: [] }
        : /^\/sessions\/[^/]+\/[^/]+$/.test(path) ? { session: SESSION }
        : {};
      return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(data) });
    });
    await context.routeWebSocket("**/api/v1/ws**", (ws) => {
      ws.send(JSON.stringify({ type: "hello", server_version: "example", hub_name: "Example hub", features: [] }));
      ws.onMessage((m) => { if (String(m).includes("app_ping")) ws.send(JSON.stringify({ type: "pong" })); });
    });
    const page = await context.newPage();
    await page.goto(base.href);
    const editor = page.locator("[data-inline-comment-editor]");
    const note = editor.locator("textarea");
    const choose = async (line, text, draft) => {
      await page.getByText(line, { exact: true }).evaluate((el, text) => {
        const node = el.firstChild;
        const start = node.textContent.indexOf(text);
        const range = new Range();
        range.setStart(node, start);
        range.setEnd(node, start + text.length);
        const selection = document.getSelection();
        selection.removeAllRanges();
        selection.addRange(range);
      }, text);
      await page.locator("[data-comment-pill]").click();
      await note.waitFor();
      assert.ok(await note.evaluate((el) => el === document.activeElement));
      assert.equal(await page.locator("[data-quotes] textarea").count(), 0);
      const mark = editor.locator("..").locator("span.absolute").last();
      const [editorRect, markRect] = await Promise.all([editor.boundingBox(), mark.boundingBox()]);
      assert.ok(Math.abs(editorRect.y - markRect.y - markRect.height) < 3);
      assert.ok(editorRect.x >= 0 && editorRect.x + editorRect.width <= width);
      await note.fill(draft);
    };
    await choose(ITEMS[1].text, "prepare", "First note");
    await page.screenshot({ path: `${output}/${width}-${theme}.png` });
    await page.evaluate(async () => {
      document.documentElement.dataset.textSize = "large";
      await new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(() => requestAnimationFrame(resolve))));
    });
    await page.screenshot({ path: `${output}/${width}-${theme}-large.png` });
    await page.evaluate(() => { delete document.documentElement.dataset.textSize; });
    await note.press("Enter");
    await choose(ITEMS[1].text, "check", "Second note");
    await note.press("Enter");
    await choose(ITEMS[3].text, "quoted passage", "Third note");
    await note.press("Enter");
    await choose(ITEMS[3].text, "your comments", "Remove me");
    await editor.getByRole("button", { name: "Remove comment", exact: true }).click();
    assert.equal(await page.locator("[data-comment-mark]").count(), 3);
    await page.locator('[data-comment-mark="2"] button').click();
    assert.equal(await note.inputValue(), "Second note");
    await note.fill("Updated second note");
    await note.press("Enter");
    const input = page.locator("[data-chat-dock] textarea").last();
    await input.fill("Review these notes.");
    const send = async (status) => {
      const response = page.waitForResponse((r) => r.url().endsWith("/messages") && r.status() === status);
      await page.locator("[data-chat-dock]").getByRole("button", { name: "Send", exact: true }).click();
      await response;
    };
    await send(400);
    const failed = page.locator("[data-comment-block]");
    await failed.nth(2).waitFor();
    assert.equal(await failed.count(), 3);
    reject = false;
    const response = page.waitForResponse((r) => r.url().endsWith("/messages") && r.status() === 200);
    await page.getByRole("button", { name: "Retry", exact: true }).click();
    await response;
    assert.equal(writes.length, 2);
    for (const { text } of writes) {
      for (const [passage, body] of [["prepare", "First note"], ["check", "Updated second note"], ["quoted passage", "Third note"]]) {
        assert.equal(text.split(body).length, 2);
        assert.ok(text.includes(`> > ${passage}\n> \n> ${body}`) || text.includes(`> > ${passage}\n>\n> ${body}`));
      }
      assert.ok(!text.includes("Remove me"));
    }
    await context.close();
    console.log(`PASS ${width}px ${theme}`);
  }
} finally {
  clearTimeout(deadline);
  await browser.close();
}
