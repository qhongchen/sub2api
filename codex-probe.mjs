import { randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const url = process.env.CODEX_URL ?? "http://100.94.253.126:8080/v1/responses";
const token = process.env.CODEX_TOKEN;
const model = process.env.CODEX_MODEL ?? "gpt-6-astra";
const includeTools = process.env.CODEX_INCLUDE_TOOLS !== "0";
const codexVersion = process.env.CODEX_VERSION ?? "0.146.0";
const reasoningEffort = process.env.CODEX_REASONING_EFFORT?.trim() || "low";

function loadInstructions() {
  if (process.env.CODEX_INSTRUCTIONS) return process.env.CODEX_INSTRUCTIONS;
  if (model.toLowerCase().includes("gpt-6-astra")) {
    try {
      return readFileSync(fileURLToPath(new URL("./backend/internal/pkg/openai/instructions_gpt6_astra.txt", import.meta.url)), "utf8");
    } catch {
      // The script can also be copied outside the repository; keep a usable fallback.
    }
  }
  return "You are Codex, a coding agent. Respond briefly to the user.";
}
const intervalMs = Number(process.env.CODEX_INTERVAL_MS ?? 60000);
const timeoutMs = Number(process.env.CODEX_TIMEOUT_MS ?? 120000);

if (!token) {
  console.error("缺少 CODEX_TOKEN");
  process.exit(1);
}

async function probe() {
  const started = Date.now();
  const requestId = randomUUID();
  const sessionId = randomUUID();
  const conversationId = sessionId;
  const body = {
    model,
    instructions: loadInstructions(),
    input: [{
      type: "message",
      role: "user",
      content: [{ type: "input_text", text: "hi" }],
    }],
    stream: true,
    store: false,
    prompt_cache_key: `channel-monitor-${model.toLowerCase().replace(/[^a-z0-9._-]+/g, "-").replace(/^-+|-+$/g, "") || "unknown"}`,
    reasoning: { effort: reasoningEffort, summary: "auto" },
    text: { verbosity: "low" },
    include: ["reasoning.encrypted_content"],
  };

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

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        authorization: `Bearer ${token}`,
        "content-type": "application/json",
        accept: "text/event-stream",
        "openai-beta": "responses=experimental",
        originator: "codex_cli_rs",
        version: codexVersion,
        session_id: sessionId,
        conversation_id: conversationId,
        "user-agent": process.env.CODEX_USER_AGENT ?? `codex_cli_rs/${codexVersion} (Ubuntu 22.4.0; x86_64) xterm-256color`,
        "x-codex-window-id": randomUUID(),
        "x-request-id": requestId,
        ...(process.env.CHATGPT_ACCOUNT_ID
          ? { "chatgpt-account-id": process.env.CHATGPT_ACCOUNT_ID }
          : {}),
      },
      body: JSON.stringify(body),
      signal: controller.signal,
    });
    const text = await response.text();
    console.log(JSON.stringify({
      time: new Date().toISOString(),
      model,
      reasoningEffort,
      includeTools,
      sessionId,
      conversationId,
      status: response.status,
      elapsedMs: Date.now() - started,
      requestId,
      upstreamRequestId: response.headers.get("x-request-id"),
      contentType: response.headers.get("content-type"),
      response: text.slice(0, 2000),
    }, null, 2));
  } catch (error) {
    console.error(JSON.stringify({
      time: new Date().toISOString(),
      model,
      reasoningEffort,
      includeTools,
      sessionId,
      conversationId,
      elapsedMs: Date.now() - started,
      requestId,
      error: error instanceof Error ? error.message : String(error),
    }, null, 2));
  } finally {
    clearTimeout(timer);
  }
}

await probe();
setInterval(probe, intervalMs);
