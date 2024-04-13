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

public class GenericSender {
    private map<ProtocolClient> clients = {};

    public function init() {
        // Read the env variable
        string? pipeConfig = env:get("OUTPUT_PIPES");
        if pipeConfig is string {
            string[] pipes = pipeConfig.split(",");
            foreach var pipe in pipes {
                string protocol = getProtocol(pipe);
                if !self.clients.hasKey(protocol) {
                    self.clients[protocol] = initializeClient(protocol, pipe);
                }
            }
        } else {
            io:println("No OUTPUT_PIPES provided");
        }
    }

    // Create client instance based on protocol and endpoint
    function initializeClient(string protocol, string endpoint) returns ProtocolClient {
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
            return {protocol: "rabbitmq", client: rabbitmqClient};
        }
        io:println("Unsupported protocol: ", protocol);
        return {};
    }

    // Send message to respective broker
    public function sendMessage(string message) {
        foreach var [protocol, client] in self.clients.entries() {
            if protocol == "kafka" {
                kafka:Producer kafkaProducer = <kafka:Producer>client.client;
                kafka:Error? result = kafkaProducer->send({topic: "test-topic", value: message});
            } else if protocol == "mqtt" {
                mqtt:Client mqttClient = <mqtt:Client>client.client;
                mqtt:Error? result = mqttClient->publishMessage({topic: "test", message: message});
            } else if protocol == "rabbitmq" {
                rabbitmq:Client rabbitmqClient = <rabbitmq:Client>client.client;
                rabbitmq:Error? result = rabbitmqClient->publishMessage({queueName: "test", message: message});
            }
        }
    }

    // Extract protocol from pipe address
    function getProtocol(string endpoint) returns string {
        if endpoint.startsWith("kafka://") {
            return "kafka";
        } else if endpoint.startsWith("mqtt://") {
            return "mqtt";
        } else if endpoint.contains("rabbitmq") {
            return "rabbitmq";
        }
        return "unknown";
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
