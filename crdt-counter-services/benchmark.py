import requests
from concurrent.futures import ThreadPoolExecutor
import time

SERVICE1 = "http://localhost:8081"
SERVICE2 = "http://localhost:8082"
OPS = 50  # number of increments per service

def increment_service(url, times):
    for _ in range(times):
        try:
            requests.get(f"{url}/increment")
        except Exception as e:
            print(f"Error incrementing {url}: {e}")

def get_value(url):
    try:
        resp = requests.get(f"{url}/value")
        resp.raise_for_status()
        value_str = resp.text.strip()
        # Parse "Value: N"
        return int(value_str.split(":")[1].strip())
    except Exception as e:
        print(f"Error fetching value from {url}: {e}")
        return -1

def main():
    print(f"Sending {OPS} increments to each service concurrently...")

    with ThreadPoolExecutor(max_workers=2) as executor:
        executor.submit(increment_service, SERVICE1, OPS)
        executor.submit(increment_service, SERVICE2, OPS)

    print("Waiting 5 seconds for CRDTs to sync...")
    time.sleep(5)

    v1 = get_value(SERVICE1)
    v2 = get_value(SERVICE2)

    print("Final values:")
    print(f"Service1: {v1}")
    print(f"Service2: {v2}")

    if v1 == v2:
        print("✅ CRDTs are synced correctly!")
    else:
        print("❌ CRDTs are NOT synced!")

if __name__ == "__main__":
    main()
