import ballerina/kafka;
import ballerina/mqtt;
import ballerina/rabbitmq;
import ballerina/io;
import ballerina/env;
import ballerina/time;

// Struct for configuration
type ProtocolClient record {
    string protocol;
    any client;
};

type QueueInfo record {
    string name;
    string address;
};

public class GenericSender {
    private map<ProtocolClient> clients = {};

    public function init() {
        string? pipeConfig = env:get("OUTPUT_PIPES");
        if pipeConfig is string {
            QueueInfo[] queues = parseOutputPipes(pipeConfig);
            foreach var queue in queues {
                string protocol = getProtocol(queue.address);
                if !self.clients.hasKey(protocol) {
                    self.clients[protocol] = initializeClient(protocol, queue.address, queue.name);
                }
            }
        } else {
            io:println("No OUTPUT_PIPES provided");
        }
    }

    function parseOutputPipes(string data) returns QueueInfo[] {
        QueueInfo[] queues = [];
        string[] entries = data.split(";");
        foreach var entry in entries {
            if (entry != "") {
                string[] parts = entry.split(",");
                string name = "";
                string address = "";
                foreach var part in parts {
                    string[] keyValue = part.split("=");
                    if (keyValue.length() == 2) {
                        if (keyValue[0] == "name") {
                            name = keyValue[1];
                        } else if (keyValue[0] == "address") {
                            address = keyValue[1];
                        }
                    }
                }
                if (name != "" && address != "") {
                    queues.push({name: name, address: address});
                }
            }
        }
        return queues;
    }

    // Create client instance based on protocol, endpoint, and queue name
    function initializeClient(string protocol, string endpoint, string queueName) returns ProtocolClient {
        if protocol == "kafka" {
            kafka:ProducerConfiguration producerConfig = {
                bootstrapServers: endpoint
            };
            kafka:Producer producer = new(producerConfig);
            return {protocol: "kafka", client: producer};
        } else if protocol == "mqtt" {
            mqtt:ClientConfiguration clientConfig = {
                url: endpoint
            };
            mqtt:Client client = new(clientConfig);
            return {protocol: "mqtt", client: client};
        } else if protocol == "rabbitmq" {
            rabbitmq:ConnectionConfiguration rabbitmqConfig = {
                host: endpoint,
                port: 5672 // Default rabbitMQ port
            };
            rabbitmq:Client rabbitmqClient = new(rabbitmqConfig);
            check rabbitmqClient->queueDeclare(queueName, durable = true, exclusive = false, autoDelete = false);
            return {protocol: "rabbitmq", client: rabbitmqClient};
        }
        io:println("Unsupported protocol: ", protocol);
        return {};
    }

    // Send message to a specific broker
    public function sendMessage(string message) {
        foreach var [protocol, client] in self.clients.entries() {
            string queueName = getQueueName(client.protocol); // Assume this function is properly implemented
            if protocol == "kafka" {
                kafka:Producer kafkaProducer = <kafka:Producer>client.client;
                kafka:Error? result = kafkaProducer->send({topic: queueName, value: message});
            } else if protocol == "mqtt" {
                mqtt:Client mqttClient = <mqtt:Client>client.client;
                mqtt:Error? result = mqttClient->publishMessage({topic: queueName, message: message});
            } else if protocol == "rabbitmq" {
                rabbitmq:Client rabbitmqClient = <rabbitmq:Client>client.client;
                rabbitmq:Error? result = rabbitmqClient->publishMessage({queueName: queueName, message: message});
            }
        }
    }

    // Extract protocol from pipe address
    function getProtocol(string endpoint) returns string {
        if endpoint.startsWith("kafka://") {
            return "kafka";
        } else if endpoint.startsWith("mqtt://") {
            return "mqtt";
        } else if (endpoint.contains("rabbitmq")) {
            return "rabbitmq";
        }
        return "unknown";
    }

    // Dummy implementation to extract queue name from the endpoint
    function getQueueName(string endpoint) returns string {
        // This needs to be implemented based on your specific URI format or other configuration details
        string[] parts = endpoint.split("/");
        return parts.length() > 1 ? parts[parts.length() - 1] : "default";
    }
}

public function main() {
    GenericSender sender = new;
    sender.init();

    // Reading interval from ENV or default to 1000ms (1 message per second)
    int interval = check int.fromString(env:get("INTERVAL") ?: "1000");
    interval = interval < 34 ? interval : 34;  // Max 30 messages per second

    int count = 0;
    while (true) {
        sender.sendMessage("Hello from Sender #" + count.toString());
        count += 1;
        time:sleep(interval);
    }
}
