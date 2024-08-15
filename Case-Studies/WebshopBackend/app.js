const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables
const interval = parseInt(process.env.interval, 10) || 5000;
const [queueAddress, queue] = process.env.internalOrdersPipe.split(',');
const messageString = process.env.data || '{"default": "Hello, World!"}';
const source = '/default/sender';
const type = process.env.eventType || "dummy.test";

// Parse the JSON message from the environment variable
let messageData;
try {
  messageData = JSON.parse(messageString);
} catch (error) {
  console.error('Invalid JSON message:', error);
  process.exit(1);
}

// Construct the CloudEvent message
const cloudEventMessage = {
  specversion: '1.0',
  id: `id-${Math.random()}`,
  source: source,
  type: type,
  time: new Date().toISOString(),
  data: messageData
};

// Connect to the RabbitMQ server
amqp.connect(queueAddress, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    channel.assertQueue(queue, {
      durable: true
    });

    // Function to send a message
    const sendMessage = () => {
      channel.sendToQueue(queue, Buffer.from(JSON.stringify(cloudEventMessage)));
      console.log("Sent: %s", JSON.stringify(cloudEventMessage));
    };

    // Set interval for sending messages
    setInterval(sendMessage, interval);
  });
});
