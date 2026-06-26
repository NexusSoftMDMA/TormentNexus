import { test, expect, describe } from "bun:test";
import { getEmbedding, getEmbeddings, isModelLoaded } from "../src/embeddings";

describe("embeddings", () => {
  test("model is not loaded initially", () => {
    expect(isModelLoaded()).toBe(false);
  });

  test("generates a 1024-dimensional embedding for a single text", async () => {
    const embedding = await getEmbedding("authentication login session bug");

    expect(embedding).toBeInstanceOf(Float32Array);
    expect(embedding.length).toBe(1024);

    // Values should be non-zero (not an empty/failed embedding)
    const hasNonZero = embedding.some((v) => v !== 0);
    expect(hasNonZero).toBe(true);
  }, 60000); // 60s timeout for first model load

  test("model is loaded after first embedding", () => {
    expect(isModelLoaded()).toBe(true);
  });

  test("similar texts have high cosine similarity", async () => {
    const emb1 = await getEmbedding("fixing the authentication bug on the login page");
    const emb2 = await getEmbedding("debugging auth issue in the sign-in screen");

    const similarity = cosineSimilarity(emb1, emb2);
    expect(similarity).toBeGreaterThan(0.7);
  }, 30000);

  test("unrelated texts have low cosine similarity", async () => {
    const emb1 = await getEmbedding("fixing the authentication bug on the login page");
    const emb2 = await getEmbedding("cooking a delicious pasta carbonara recipe");

    const similarity = cosineSimilarity(emb1, emb2);
    expect(similarity).toBeLessThan(0.6);
  }, 30000);

  test("batch embedding returns correct number of results", async () => {
    const texts = [
      "first text about auth",
      "second text about payments",
      "third text about deployment",
    ];

    const embeddings = await getEmbeddings(texts);

    expect(embeddings.length).toBe(3);
    for (const emb of embeddings) {
      expect(emb).toBeInstanceOf(Float32Array);
      expect(emb.length).toBe(1024);
    }
  }, 30000);

  test("empty string returns valid embedding", async () => {
    const embedding = await getEmbedding("");

    expect(embedding).toBeInstanceOf(Float32Array);
    expect(embedding.length).toBe(1024);
  }, 30000);
});

// Helper
function cosineSimilarity(a: Float32Array, b: Float32Array): number {
  let dotProduct = 0;
  let normA = 0;
  let normB = 0;
  for (let i = 0; i < a.length; i++) {
    dotProduct += a[i] * b[i];
    normA += a[i] * a[i];
    normB += b[i] * b[i];
  }
  return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
}
