// Quick health check
const http = require("http");
const req = http.get("http://127.0.0.1:4000/health", (res) => {
	let d = "";
	res.on("data", (c) => (d += c));
	res.on("end", () => {
		console.log("OK:", d);
		process.exit(0);
	});
});
req.on("error", (e) => {
	console.log("FAIL:", e.message);
	process.exit(1);
});
req.setTimeout(3000, () => {
	console.log("TIMEOUT");
	process.exit(2);
});
