import requests
from rich.console import Console
import concurrent.futures
import time
import os

console = Console()
outputs = []

def get_services_list():
    eas = os.environ.get("EARNINGS", "").split(",")
    acs = os.environ.get("ACS", "").split(",")
    prs = os.environ.get("PRS", "").split(",")
        
    return [eas, acs, prs]

services = get_services_list()


def hit_endpoint(url):
    response = requests.get(url)
    return response.json()

def check_service(service):
    service_name, service_url = service

    info = f"{service_url}/info"
    health = f"{service_url}/health"

    info = hit_endpoint(info)['git']['commit']['id']
    health = hit_endpoint(health)['status']

    if health == "UP":
        outputs.append(f"{service_name} [bold white on green] {health} [/bold white on green] [black on white] {info} [/black on white]")
    else:
        outputs.append(f"{service_name} [bold white on red] {health} [/bold white on red] [black on white] {info} [/black on white]")
        
def main():
    # Threadding this operation makes it ~65% faster than iterating.
    tick = time.perf_counter()
    with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
        executor.map(check_service, services, timeout=5)

    outputs.sort()

    for output in outputs:
        console.print(output)
    
    tock = time.perf_counter()
    console.print(f"\n:timer_clock:  {tock - tick:0.4f}s")


if __name__ == "__main__":
    main()
