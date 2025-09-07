import requests
import threading
import time
import random
import string
from concurrent.futures import ThreadPoolExecutor

# ANSI green
GREEN = "\033[92m"
RESET = "\033[0m"

# Node URLs
GCOUNTER_NODES = ["http://localhost:8081", 
                  "http://localhost:8082"]

PNCOUNTER_NODES = ["http://localhost:8083", 
                   "http://localhost:8084"]

GSET_NODES = ["http://localhost:8085", 
              "http://localhost:8086"]

ORSET_NODES = ["http://localhost:8087", 
               "http://localhost:8088"]

LWW_NODES = ["http://localhost:8089", 
             "http://localhost:8090"]

# Number of operations per node
OPERATIONS = 5
SYNC_WAIT = 1  # seconds to wait between operations

def increment_node(url):
    try:
        requests.post(url + "/inc")
    except Exception as e:
        print(f"Error incrementing {url}: {e}")

def decrement_node(url):
    try:
        requests.post(url + "/dec")
    except Exception as e:
        print(f"Error decrementing {url}: {e}")

def increment_or_decrement_node(url):
    if random.choice([True, False]):
        increment_node(url)
    else:
        decrement_node(url)

def add_random_element(url):
    element = ''.join(random.choices(string.ascii_letters, k=3))
    try:
        requests.post(url + f"/add?element={element}")
    except Exception as e:
        print(f"Error adding element to {url}: {e}")

def get_value(url):
    try:
        r = requests.get(url + "/value")
        return int(r.text.strip().split(":")[-1])
    except Exception as e:
        print(f"Error fetching value from {url}: {e}")
        return None

def get_state(url):
    try:
        r = requests.get(url + "/state")
        return r.json()
    except Exception as e:
        print(f"Error fetching state from {url}: {e}")
        return None

def benchmark_counter(nodes, name="Counter"):
    print(f"\nBenchmarking {name}...")
    for i in range(OPERATIONS):
        with ThreadPoolExecutor(max_workers=len(nodes)) as executor:
            for node in nodes:
                executor.submit(increment_node, node)
                
        time.sleep(SYNC_WAIT)
        values = [get_value(node) for node in nodes]
        print(f"Step {i+1}: {values}")

    time.sleep(3* SYNC_WAIT)
    final_values = [get_value(node) for node in nodes]
    print(f"{GREEN}Final {name} node values: {final_values}{RESET}")
    print(f"{GREEN}[{name}] All equal? {all(v == final_values[0] for v in final_values)}{RESET}")

def benchmark_pncounter(nodes, name="PNCounter"):
    print(f"\nBenchmarking {name}...")
    for i in range(OPERATIONS):
        with ThreadPoolExecutor(max_workers=len(nodes)) as executor:
            for node in nodes:
                executor.submit(increment_or_decrement_node, node)
                
        time.sleep(SYNC_WAIT)
        values = [get_value(node) for node in nodes]
        print(f"Step {i+1}: {values}")

    time.sleep(3 *SYNC_WAIT)
    final_values = [get_value(node) for node in nodes]
    print(f"{GREEN}Final {name} node values: {final_values}{RESET}")
    print(f"{GREEN}[{name}] All equal? {all(v == final_values[0] for v in final_values)}{RESET}")

def benchmark_gset(nodes, name="Set"):
    print(f"\nBenchmarking {name}...")
    for i in range(OPERATIONS):
        with ThreadPoolExecutor(max_workers=len(nodes)) as executor:
            for node in nodes:
                executor.submit(add_random_element, node)
        time.sleep(SYNC_WAIT)
        states = [get_state(node) for node in nodes]
        print(f"Step {i+1}: {states}")

    time.sleep(3*SYNC_WAIT)
    final_states = [get_state(node) for node in nodes]
    print(f"{GREEN}Final {name} node states: {final_states}{RESET}")

    # Check equality
    sets = [set(s) if s else set() for s in final_states]
    all_equal = all(s == sets[0] for s in sets)
    print(f"{GREEN}[{name}] All equal? {all_equal}{RESET}")

def remove_random_element(url):
    """Try to remove a random element from the ORSet node."""
    state = get_state(url)
    if state:
        elem = random.choice(state)
        try:
            requests.post(url + f"/remove?element={elem}")
        except Exception as e:
            print(f"Error removing element from {url}: {e}")

def benchmark_orset(nodes, name="ORSet"):
    print(f"\nBenchmarking {name}...")
    for i in range(OPERATIONS):
        with ThreadPoolExecutor(max_workers=len(nodes)) as executor:
            for node in nodes:
                if random.random() < 0.3:  # 30% chance to remove
                    executor.submit(remove_random_element, node)
                else:
                    executor.submit(add_random_element, node)
        time.sleep(SYNC_WAIT)
        states = [get_state(node) for node in nodes]
        print(f"Step {i+1}: {states}")

    time.sleep(3*SYNC_WAIT)
    final_states = [get_state(node) for node in nodes]
    print(f"{GREEN}Final {name} node states: {final_states}{RESET}")

    sets = [set(s) if s else set() for s in final_states]
    all_equal = all(s == sets[0] for s in sets)
    print(f"{GREEN}[{name}] All equal? {all_equal}{RESET}")

def set_random_lww(url):
    value = ''.join(random.choices(string.ascii_letters + string.digits, k=5))
    try:
        requests.post(url + f"/set?value={value}")
    except Exception as e:
        print(f"Error setting LWW value on {url}: {e}")
    return value

def get_lww_value(url):
    try:
        r = requests.get(url + "/value")
        # The response format is "Value: XYZ", so split by ':' and strip
        return r.text.strip().split(":")[-1].strip()
    except Exception as e:
        print(f"Error fetching LWW value from {url}: {e}")
        return None

def benchmark_lwwregister(nodes, name="LWWRegister"):
    print(f"\nBenchmarking {name}...")
    for i in range(OPERATIONS):
        with ThreadPoolExecutor(max_workers=len(nodes)) as executor:
            executor.submit(set_random_lww, nodes[0])
            executor.submit(set_random_lww, nodes[1])
        time.sleep(SYNC_WAIT)
        values = [get_lww_value(node) for node in nodes]
        print(f"Step {i+1}: {values}")

    time.sleep(3*SYNC_WAIT)

    final_values = [get_lww_value(node) for node in nodes]
    all_equal = all(v == final_values[0] for v in final_values)
    print(f"{GREEN}Final {name} node values: {final_values}{RESET}")
    print(f"{GREEN}[{name}] All equal? {all_equal}{RESET}")
         
if __name__ == "__main__":
    benchmark_counter(GCOUNTER_NODES, "GCounter")
    benchmark_pncounter(PNCOUNTER_NODES, "PNCounter")
    benchmark_gset(GSET_NODES, "GSet")
    benchmark_orset(ORSET_NODES, "ORSet")
    benchmark_lwwregister(LWW_NODES, "LWWRegister")


    # Small sleep to let all threads be cleaned up properly
    time.sleep(1)
