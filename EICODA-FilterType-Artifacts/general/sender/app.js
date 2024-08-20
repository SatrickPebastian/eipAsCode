const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables
const interval = parseInt(process.env.interval, 10);
const [queueAddress, queue] = process.env.out.split(',');
const messageString = process.env.data;
const source = process.env.source;
const type = process.env.eventType;

let messageData;
try {
  messageData = JSON.parse(messageString);
} catch (error) {
  console.error('Invalid JSON message:', error);
  process.exit(1);
}

// Build CloudEvent
const cloudEventMessage = {
  specversion: '1.0',
  id: `id-${Math.random()}`,
  source: source,
  type: type,
  time: new Date().toISOString(),
  data: messageData
};

amqp.connect(queueAddress, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    channel.assertQueue(queue);

    const sendMessage = () => {
      channel.sendToQueue(queue, Buffer.from(JSON.stringify(cloudEventMessage)));
      console.log("Sent: %s", JSON.stringify(cloudEventMessage));
    };

    //Set interval for sending messages
    setInterval(sendMessage, interval);
  });
});
