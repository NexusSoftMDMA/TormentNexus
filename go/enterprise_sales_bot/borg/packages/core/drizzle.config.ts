import { defineConfig } from "drizzle-kit";

export default defineConfig({
    schema: "./src/db/index.ts",
    out: "./drizzle",
    dialect: "sqlite",
    dbCredentials: { url: "./tormentnexus.db" },
});
