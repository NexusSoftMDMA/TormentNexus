const OLLAMA_URL = "http://localhost:11434";
const OLLAMA_MODEL = "qwen3-embedding:0.6b";
export const EMBEDDING_DIMS = 1024;

let ollamaChecked = false;

export function isModelLoaded(): boolean {
  return ollamaChecked;
}

/**
 * Check if Ollama is running and has the embedding model.
 */
async function pullModel(): Promise<void> {
  console.error(`[claude-find] Model '${OLLAMA_MODEL}' not found — pulling automatically...`);
  const resp = await fetch(`${OLLAMA_URL}/api/pull`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ model: OLLAMA_MODEL }),
  });
  if (!resp.ok || !resp.body) {
    throw new Error(`Failed to pull model: ${resp.status} ${resp.statusText}`);
  }
  // Consume the streaming response until complete
  const reader = resp.body.getReader();
  const decoder = new TextDecoder();
  let lastStatus = "";
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    const text = decoder.decode(value, { stream: true });
    for (const line of text.split("\n").filter(Boolean)) {
      try {
        const obj = JSON.parse(line);
        if (obj.status && obj.status !== lastStatus) {
          console.error(`[claude-find] Pull: ${obj.status}`);
          lastStatus = obj.status;
        }
      } catch {}
    }
  }
  console.error(`[claude-find] Model '${OLLAMA_MODEL}' pulled successfully.`);
}

async function checkOllama(): Promise<void> {
  if (ollamaChecked) return;

  try {
    const resp = await fetch(`${OLLAMA_URL}/api/tags`, { signal: AbortSignal.timeout(2000) });
    if (!resp.ok) throw new Error("Ollama not responding");
    const data = await resp.json() as any;
    const models = data.models || [];
    const hasModel = models.some((m: any) => m.name?.startsWith(OLLAMA_MODEL));
    if (!hasModel) {
      await pullModel();
    }
    ollamaChecked = true;
    console.error("[claude-find] Using Ollama (Metal GPU accelerated)");
  } catch (err) {
    console.error("[claude-find] Error: Ollama is required but not available.");
    console.error("[claude-find] Install: brew install ollama && brew services start ollama");
    throw new Error(`Ollama required: ${err instanceof Error ? err.message : String(err)}`);
  }
}

/**
 * Embed via Ollama HTTP API.
 * Qwen3-Embedding uses instruction prefixes for queries, raw text for documents.
 */
async function embedViaOllama(texts: string[], mode: "document" | "query"): Promise<Float32Array[]> {
  const input = mode === "query"
    ? texts.map((t) => `Instruct: Given a search query, retrieve relevant conversation passages\nQuery: ${t}`)
    : texts;
  const resp = await fetch(`${OLLAMA_URL}/api/embed`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ model: OLLAMA_MODEL, input, keep_alive: -1 }),
  });

  if (!resp.ok) {
    throw new Error(`Ollama embed failed: ${resp.status} ${resp.statusText}`);
  }

  const data = await resp.json() as any;
  return data.embeddings.map((emb: number[]) => new Float32Array(emb));
}

/**
 * Embed a single text.
 */
export async function getEmbedding(
  text: string,
  mode: "document" | "query" = "document"
): Promise<Float32Array> {
  await checkOllama();
  const [result] = await embedViaOllama([text], mode);
  return result;
}

/**
 * Embed multiple texts in a batch.
 */
export async function getEmbeddings(
  texts: string[],
  mode: "document" | "query" = "document"
): Promise<(Float32Array | null)[]> {
  await checkOllama();

  const BATCH_SIZE = 8;
  const results: (Float32Array | null)[] = [];
  for (let i = 0; i < texts.length; i += BATCH_SIZE) {
    const batch = texts.slice(i, i + BATCH_SIZE);
    try {
      const batchResults = await embedViaOllama(batch, mode);
      results.push(...batchResults);
    } catch (err: any) {
      // Only retry individually for content errors (4xx)
      const status = err?.message?.match(/(\d{3})/)?.[1];
      if (status && parseInt(status) >= 400 && parseInt(status) < 500) {
        for (const text of batch) {
          try {
            const [single] = await embedViaOllama([text], mode);
            results.push(single);
          } catch {
            console.error(`[claude-find] Warning: failed to embed chunk (${text.length} chars), skipping`);
            results.push(null);
          }
        }
      } else {
        throw err;
      }
    }
  }
  return results;
}
