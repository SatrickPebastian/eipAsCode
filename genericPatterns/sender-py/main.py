from dapr.clients import DaprClient
import os
import time
import json

class GenericSender:
    def __init__(self):
        self.queues = []

    def init(self):
        pipe_config = os.getenv('OUTPUT_PIPES')
        if pipe_config:
            self.queues = self.parse_output_pipes(pipe_config)
        else:
            print("No OUTPUT_PIPES provided")

    def parse_output_pipes(self, data):
        return [dict(zip(['name', 'address'], entry.split(','))) for entry in data.split(';') if entry]

    def send_message(self, message):
        with DaprClient() as client:
            for queue_info in self.queues:
                binding_name = queue_info['address'].split('://')[0] 
                client.invoke_binding(binding_name, 'create', json.dumps({'data': message}))

if __name__ == "__main__":
    sender = GenericSender()
    sender.init()

    interval = int(os.getenv('INTERVAL', "1000"))
    interval = max(34, interval)  # Max 30 messages per second

    count = 0
    while True:
        sender.send_message(f"Hello from Sender #{count}")
        count += 1
        time.sleep(interval / 1000.0)
