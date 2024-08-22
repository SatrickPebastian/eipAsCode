const amqp = require('amqplib/callback_api');
const process = require('process');

const interval = parseInt(process.env.interval, 10);
const [pipeAddress, pipe, pipeType] = process.env.out.split(',');
const outRoutingKey = process.env.outRoutingKey || '';
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

amqp.connect(pipeAddress, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    if (pipeType === 'queue') {
      channel.assertQueue(pipe);

      const sendMessage = () => {
        // Create a new CloudEvent message for each call
        const cloudEventMessage = {
          specversion: '1.0',
          id: `id-${Math.random()}`,
          source: source,
          type: type,
          time: new Date().toISOString(),
          data: messageData
        };

        channel.sendToQueue(pipe, Buffer.from(JSON.stringify(cloudEventMessage)));
        console.log("Sent to queue %s: %s", pipe, JSON.stringify(cloudEventMessage));
      };

      setInterval(sendMessage, interval);
    } else if (pipeType === 'topic') {
      channel.assertExchange(pipe, 'topic');

      const sendMessage = () => {
        // Create a new CloudEvent message
        const cloudEventMessage = {
          specversion: '1.0',
          id: `id-${Math.random()}`,
          source: source,
          type: type,
          time: new Date().toISOString(),
          data: messageData
        };

        channel.publish(pipe, outRoutingKey, Buffer.from(JSON.stringify(cloudEventMessage)));
        console.log("Sent to exchange %s with routing key %s: %s", pipe, outRoutingKey, JSON.stringify(cloudEventMessage));
      };

      setInterval(sendMessage, interval);
    } else {
      console.error(`Unknown pipe type: ${pipeType}`);
      process.exit(1);
    }
  });
});
