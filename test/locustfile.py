import uuid
import time
from locust import HttpUser, task


class HelloWorldUser(HttpUser):
    host = "http://localhost:9999"

    # @task
    # def summary(self):
    #     self.client.get("/payments-summary")
    #     time.sleep(10)

    # @task
    # def payment(self):
    #     id = str(uuid.uuid4())
    #     self.client.post("/payments", json={"correlationId": id, "amount": 100})
    #     time.sleep(0.1)

    @task
    def payment(self):
        self.client.get("/payments-summary")
        time.sleep(0.1)
