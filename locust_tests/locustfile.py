from locust import HttpUser, task, between, tag
import random, itertools

class ReverseProxyUser(HttpUser):
    wait_time = between(1, 5)
    upstreams = ['stub-server-1', 'stub-server-2']

    @task
    def send_request(self):
        selected_upstream = random.choice(self.upstreams)
        print(f"Sending request to upstream: {selected_upstream}")
        headers = {'X-Upstream': selected_upstream}
        self.client.get("/", headers=headers)
