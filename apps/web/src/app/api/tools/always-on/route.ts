import { type NextRequest, NextResponse } from "next/server";
import { readFileSync, writeFileSync, existsSync } from "fs";
import { join } from "path";

const WORKSPACE_DATA_DIR = existsSync(join(process.cwd(), "..", "..", "data"))
	? join(process.cwd(), "..", "..", "data")
	: join(process.cwd(), "data");

const CONFIG_PATH = join(WORKSPACE_DATA_DIR, "always-on-tools.json");

interface AlwaysOnConfig {
	tools: Record<string, boolean>;
}

function loadConfig(): AlwaysOnConfig {
	try {
		if (existsSync(CONFIG_PATH)) {
			return JSON.parse(readFileSync(CONFIG_PATH, "utf-8"));
		}
	} catch {
		// ignore
	}
	return { tools: {} };
}

function saveConfig(config: AlwaysOnConfig) {
	if (!existsSync(WORKSPACE_DATA_DIR)) {
		const fs = require("fs");
		fs.mkdirSync(WORKSPACE_DATA_DIR, { recursive: true });
	}
	writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2), "utf-8");
}

export async function POST(request: NextRequest) {
	try {
		const { name, alwaysOn } = await request.json();

		if (!name || typeof name !== "string") {
			return NextResponse.json(
				{ success: false, error: "Missing or invalid 'name'" },
				{ status: 400 },
			);
		}

		const config = loadConfig();
		config.tools[name] = alwaysOn === true;
		saveConfig(config);

		return NextResponse.json({
			success: true,
			name,
			alwaysOn: config.tools[name],
		});
	} catch (error) {
		const message = error instanceof Error ? error.message : "Unknown error";
		return NextResponse.json(
			{ success: false, error: message },
			{ status: 500 },
		);
	}
}
