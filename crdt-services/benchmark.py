import requests
import time
import statistics
import matplotlib.pyplot as plt

SERVICES = {
    "pn_counter": "http://localhost:8081",
    "gset": "http://localhost:8082",
    "orset": "http://localhost:8083",
    "lww_register": "http://localhost:8084"
}

def benchmark_service(name, base_url, ops=100):
    latencies = []

    if name == "pn_counter":
        # increment
        for i in range(ops):
            start = time.time()
            requests.post(f"{base_url}/increment", json={"n": 1})
            r = requests.get(f"{base_url}/value")
            end = time.time()
            latencies.append(end - start)
        final_val = r.json()["value"]

    elif name == "gset":
        for i in range(ops):
            start = time.time()
            requests.post(f"{base_url}/add", json={"element": f"e{i}"})
            r = requests.get(f"{base_url}/elements")
            end = time.time()
            latencies.append(end - start)
        final_val = len(r.json()["elements"])

    elif name == "orset":
        for i in range(ops):
            start = time.time()
            requests.post(f"{base_url}/add", json={"element": f"x{i}"})
            r = requests.get(f"{base_url}/elements")
            end = time.time()
            latencies.append(end - start)
        final_val = len(r.json()["elements"])

    elif name == "lww_register":
        for i in range(ops):
            start = time.time()
            requests.post(f"{base_url}/set", json={"value": f"v{i}"})
            r = requests.get(f"{base_url}/value")
            end = time.time()
            latencies.append(end - start)
        final_val = r.json()["value"]

    return {
        "service": name,
        "ops": ops,
        "mean_latency": statistics.mean(latencies),
        "p95_latency": statistics.quantiles(latencies, n=100)[94],
        "final_value": final_val
    }

def main():
    results = []
    for name, url in SERVICES.items():
        print(f"Benchmarking {name}...")
        result = benchmark_service(name, url, ops=100)
        results.append(result)
        print(result)

    # Plotting results
    names = [r["service"] for r in results]
    mean_latencies = [r["mean_latency"] for r in results]
    p95_latencies = [r["p95_latency"] for r in results]

    plt.bar(names, mean_latencies)
    plt.title("Mean Latency per CRDT")
    plt.ylabel("Seconds")
    plt.show()

    plt.bar(names, p95_latencies)
    plt.title("P95 Latency per CRDT")
    plt.ylabel("Seconds")
    plt.show()

if __name__ == "__main__":
    main()
