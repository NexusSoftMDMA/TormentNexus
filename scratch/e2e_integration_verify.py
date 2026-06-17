import requests
import time
import json

BASE_URL = "http://localhost:4300"

def test_connectivity():
    print("--- Testing Service Connectivity ---")
    try:
        resp = requests.get(f"{BASE_URL}/api/service/connectivity")
        print(f"Status: {resp.status_code}")
        print(resp.json())
        return resp.status_code == 200
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_ripgrep():
    print("\n--- Testing Native Tool: ripgrep_search ---")
    payload = {
        "name": "ripgrep_search",
        "arguments": {"pattern": "package", "path": "."}
    }
    try:
        start = time.time()
        resp = requests.post(f"{BASE_URL}/api/agent/tool", json=payload)
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            print("✅ Ripgrep execution verified.")
            return True
        else:
            print(f"❌ Ripgrep failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_skills():
    print("\n--- Testing Skill Registry ---")
    try:
        start = time.time()
        resp = requests.get(f"{BASE_URL}/api/skills")
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            skills = resp.json().get("data", [])
            print(f"✅ Listed {len(skills)} skills via /api/skills.")
            return True
        else:
            print(f"❌ Skills failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_scripts():
    print("\n--- Testing Prompt Library ---")
    try:
        start = time.time()
        resp = requests.get(f"{BASE_URL}/api/scripts")
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            print("✅ API access to scripts/prompts verified.")
            return True
        else:
            print(f"❌ Scripts failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_system_overview():
    print("\n--- Testing Memory Tracking Schema ---")
    try:
        start = time.time()
        resp = requests.get(f"{BASE_URL}/api/system/overview")
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            print("✅ System overview is healthy.")
            return True
        else:
            print(f"❌ System overview failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

if __name__ == "__main__":
    print("🚀 TormentNexus E2E Integration Protocol v1\n")
    all_pass = True
    all_pass &= test_connectivity()
    all_pass &= test_ripgrep()
    all_pass &= test_skills()
    all_pass &= test_scripts()
    all_pass &= test_system_overview()

    if all_pass:
        print("\n🏁 Integration Tests Complete: ALL PASSED")
    else:
        print("\n🏁 Integration Tests Complete: FAILED")
        exit(1)
import time
import urllib.request
import json
import os

BASE_URL = "http://localhost:4300"
SAMPLE_REPO = "/tmp/tormentnexus-sample"

def call_endpoint(path, method='GET', payload=None):
    url = f"{BASE_URL}{path}"
    data = json.dumps(payload).encode('utf-8') if payload else None
    headers = {'Content-Type': 'application/json'} if payload else {}
    req = urllib.request.Request(url, data=data, method=method, headers=headers)

    print(f"\n>>> {method} {path}")
    if payload: print(f"Payload: {json.dumps(payload)}")

    start = time.perf_counter()
    try:
        with urllib.request.urlopen(req) as response:
            res_body = response.read().decode('utf-8')
            end = time.perf_counter()
            duration = (end - start) * 1000
            result = json.loads(res_body)
            print(f"Status: {response.status} | Latency: {duration:.2f}ms")
            return result
    except Exception as e:
        print(f"Status: Failed | Error: {e}")
        return None

def call_tool(name, arguments):
    return call_endpoint("/api/agent/tool", "POST", {"name": name, "arguments": arguments})

def run_e2e():
    print("🚀 TormentNexus E2E Integration Protocol v1\n")

    # 1. Native Tool Execution: Ripgrep
    print("--- Testing Native Tool: ripgrep_search ---")
    rg_res = call_tool("ripgrep_search", {"pattern": "TormentNexus", "path": SAMPLE_REPO})
    if rg_res and "success" in rg_res:
        print("✅ Ripgrep execution verified.")

    # 2. Skill Registry Operations
    print("\n--- Testing Skill Registry ---")
    skills = call_endpoint("/api/skills")
    if skills and "success" in skills:
        data = skills.get('data', [])
        print(f"✅ Listed {len(data)} skills via /api/skills.")

    # 3. Prompt Library Operations
    print("\n--- Testing Prompt Library ---")
    prompts = call_endpoint("/api/scripts") #saved scripts
    if prompts and "success" in prompts:
        print("✅ API access to scripts/prompts verified.")

    # 4. Memory Tracking & De-duplication Table Check
    print("\n--- Testing Memory Tracking Schema ---")
    overview = call_endpoint("/api/system/overview")
    if overview and "success" in overview:
        print("✅ System overview is healthy.")

    print("\n🏁 Integration Tests Complete")

if __name__ == "__main__":
    run_e2e()
