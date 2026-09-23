#!/usr/bin/env node

const url = process.env.CODEX_URL || "https://chatgpt.com/backend-api/codex/responses";
const token = process.env.CODEX_TOKEN;
const accountId = process.env.CHATGPT_ACCOUNT_ID || "";
const model = process.env.CODEX_MODEL || "gpt-6-astra";
const intervalMs = Number(process.env.CODEX_INTERVAL_MS || 60_000);
const timeoutMs = Number(process.env.CODEX_TIMEOUT_MS || 90_000);
const includeTools = process.env.CODEX_INCLUDE_TOOLS !== "0";

if (!token) {
  console.error("缺少 CODEX_TOKEN");
  process.exit(1);
}

const sessionId = crypto.randomUUID();
const conversationId = crypto.randomUUID();

const instructions = process.env.CODEX_INSTRUCTIONS ||
  "You are Codex. Answer the user's request directly and briefly.";

function buildBody() {
  const body = {
    model,
    instructions,
    input: [{
      type: "message",
      role: "user",
      content: [{ type: "input_text", text: "hi" }],
    }],
    stream: true,
    store: false,
    reasoning: { effort: "low", summary: "auto" },
    text: { verbosity: "low" },
    include: ["reasoning.encrypted_content"],
  };

  // 默认打开，便于和真实 Codex 请求比较。
  // 只想测试无工具请求时设置 CODEX_INCLUDE_TOOLS=0。
  if (includeTools) {
    body.tools = [{
      type: "function",
      name: "shell",
      description: "Run a shell command",
      parameters: {
        type: "object",
        properties: { cmd: { type: "string" } },
        required: ["cmd"],
      },
      strict: false,
    }];
    body.tool_choice = "auto";
    body.parallel_tool_calls = true;
  }

  return body;
}

async function probe(index) {
  const started = Date.now();
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  const headers = {
    authorization: `Bearer ${token}`,
    "content-type": "application/json",
    accept: "text/event-stream",
    originator: "codex_cli_rs",
    version: process.env.CODEX_VERSION || "0.1.0",
    session_id: sessionId,
    conversation_id: conversationId,
    "user-agent": process.env.CODEX_USER_AGENT || "codex_cli_rs/0.1.0",
  };

  if (accountId) headers["chatgpt-account-id"] = accountId;

  try {
    const response = await fetch(url, {
      method: "POST",
      headers,
      body: JSON.stringify(buildBody()),
      signal: controller.signal,
    });

    const raw = await response.text();
    const elapsed = Date.now() - started;
    const requestId =
      response.headers.get("x-request-id") ||
      response.headers.get("x-openai-request-id") ||
      "";

    const events = raw
      .split("\\n")
      .filter(line => line.startsWith("data:"))
      .map(line => line.slice(5).trim())
      .filter(Boolean)
      .filter(value => value !== "[DONE]")
      .slice(-8)
      .map(value => {
        try {
          const json = JSON.parse(value);
          return json.type || "json";
        } catch {
          return value.slice(0, 120);
        }
      });

    console.log(JSON.stringify({
      time: new Date().toISOString(),
      index,
      model,
      includeTools,
      status: response.status,
      elapsedMs: elapsed,
      requestId,
      contentType: response.headers.get("content-type"),
      events,
      body: response.ok ? undefined : raw.slice(0, 2000),
    }));
  } catch (error) {
    console.error(JSON.stringify({
      time: new Date().toISOString(),
      index,
      model,
      includeTools,
      elapsedMs: Date.now() - started,
      error: error.name === "AbortError" ? "timeout" : error.message,
    }));
  } finally {
    clearTimeout(timer);
  }
}

let index = 0;

while (true) {
  await probe(++index);
  await new Promise(resolve => setTimeout(resolve, intervalMs));
}
