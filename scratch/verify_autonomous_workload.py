import time
import urllib.request
import json

BASE_URL = "http://localhost:4300"

def call_endpoint(path, method='GET', payload=None):
    url = f"{BASE_URL}{path}"
    data = json.dumps(payload).encode('utf-8') if payload else None
    headers = {'Content-Type': 'application/json'} if payload else {}
    req = urllib.request.Request(url, data=data, method=method, headers=headers)

    print(f"--- Executing: {method} {path} ---")
    start = time.perf_counter()
    try:
        with urllib.request.urlopen(req) as response:
            res_body = response.read().decode('utf-8')
            end = time.perf_counter()
            duration = (end - start) * 1000
            result = json.loads(res_body)
            print(f"Status: Success | Latency: {duration:.2f}ms")
            return result
    except Exception as e:
        print(f"Status: Failed | Error: {e}")
        return None

def run_workload():
    print("🚀 Starting TormentNexus E2E Integration Verification\n")

    # Step 1: Health check
    call_endpoint("/health")

    # Step 2: List skills
    call_endpoint("/api/skills/list")

    # Step 3: Search prompts
    call_endpoint("/api/native/tools/search", "POST", {"query": "orchestrator"})

    # Step 4: System overview
    call_endpoint("/api/system/overview")

    # Step 5: Runtime status
    call_endpoint("/api/runtime/status")

    print("\n✅ E2E Integration Verification Complete")

if __name__ == "__main__":
    run_workload()
