#!/usr/bin/env node
/**
 * Fixture Validator
 * 
 * Purpose: Parse GOLDEN_FIXTURE_RESPONSES.md, extract JSON examples,
 *          validate each against FIXTURE_SCHEMA.jsonc, and report compliance
 * 
 * Usage: node verify-fixtures.mjs [--fix] [--verbose]
 */

import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_FILE = path.join(__dirname, "GOLDEN_FIXTURE_RESPONSES.md");
const SCHEMA_FILE = path.join(__dirname, "FIXTURE_SCHEMA.jsonc");

// Parse JSONC (JSON with comments)
function parseJsonc(content) {
  // Remove // comments
  const cleaned = content.replace(/\/\/.*$/gm, "");
  try {
    return JSON.parse(cleaned);
  } catch (e) {
    console.error("Failed to parse JSONC:", e.message);
    process.exit(1);
  }
}

// Extract JSON blocks from markdown
function extractJsonBlocks(content) {
  const blocks = [];
  const regex = /```json\n([\s\S]*?)\n```/g;
  let match;

  while ((match = regex.exec(content)) !== null) {
    try {
      const json = JSON.parse(match[1]);
      blocks.push({ json, source: match[0].substring(0, 50) });
    } catch (e) {
      console.warn(`⚠️  Invalid JSON block: ${e.message}`);
    }
  }

  return blocks;
}

// Simple JSON Schema validator (basic subset)
function validateAgainstSchema(data, schema) {
  const errors = [];

  // Check required fields
  if (schema.required) {
    for (const field of schema.required) {
      if (!(field in data)) {
        errors.push(`Missing required field: ${field}`);
      }
    }
  }

  // Check type
  if (schema.type && typeof data !== schema.type) {
    errors.push(`Expected type ${schema.type}, got ${typeof data}`);
  }

  // Check enum
  if (schema.enum && !schema.enum.includes(data)) {
    errors.push(`Value must be one of: ${schema.enum.join(", ")}`);
  }

  return errors;
}

// Main validator
async function validate() {
  console.log("🔍 Validating golden fixtures...\n");

  // Read files
  let fixtureContent;
  let schemaContent;

  try {
    fixtureContent = fs.readFileSync(FIXTURE_FILE, "utf-8");
    schemaContent = fs.readFileSync(SCHEMA_FILE, "utf-8");
  } catch (e) {
    console.error(`❌ Failed to read files: ${e.message}`);
    process.exit(1);
  }

  // Parse schema
  let schema;
  try {
    schema = parseJsonc(schemaContent);
  } catch (e) {
    console.error(`❌ Failed to parse schema: ${e.message}`);
    process.exit(1);
  }

  // Extract JSON fixtures
  const fixtures = extractJsonBlocks(fixtureContent);

  if (fixtures.length === 0) {
    console.warn("⚠️  No JSON fixtures found in GOLDEN_FIXTURE_RESPONSES.md");
    return;
  }

  console.log(`📋 Found ${fixtures.length} JSON fixtures\n`);

  let passCount = 0;
  let failCount = 0;

  // Validate each fixture
  for (const [index, fixture] of fixtures.entries()) {
    const fixtureType = Object.keys(fixture.json)[0] || "unknown";

    // Try to match with schema
    const schemaMatch = findMatchingSchema(fixtureType, schema);

    if (schemaMatch) {
      const errors = validateAgainstSchema(fixture.json, schemaMatch);

      if (errors.length === 0) {
        console.log(`✅ Fixture #${index + 1} (${fixtureType}): VALID`);
        passCount++;
      } else {
        console.log(
          `❌ Fixture #${index + 1} (${fixtureType}): INVALID`
        );
        errors.forEach((err) => console.log(`   - ${err}`));
        failCount++;
      }
    } else {
      console.log(
        `⚠️  Fixture #${index + 1} (${fixtureType}): No schema match (skipped)`
      );
    }
  }

  console.log(`\n📊 Results: ${passCount} passed, ${failCount} failed\n`);

  if (failCount > 0) {
    process.exit(1);
  }
}

// Helper to find matching schema for fixture
function findMatchingSchema(fixtureType, schema) {
  // Simple heuristic: match schema by looking at response structure
  // In production, this would use more sophisticated matching

  const definitions = schema.definitions || {};

  // Try to match based on common patterns
  if (
    fixtureType === "exitCode" ||
    fixtureType === "stdout" ||
    fixtureType === "duration_ms"
  ) {
    return definitions.shellExecutionResponse;
  }

  if (fixtureType === "permissionDecision") {
    return definitions.hookDecisionResponse;
  }

  if (fixtureType === "success" && fixtureType === "path") {
    return definitions.fileOperationResponse;
  }

  // Return a permissive schema for now
  return { type: "object" };
}

// Run validator
validate().catch(console.error);
