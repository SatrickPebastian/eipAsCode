const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables
const interval = parseInt(process.env.INTERVAL, 10) || 5000;
const queueAddress = process.env.QUEUE_ADDRESS || 'amqp://guest:guest@localhost:5672';
const message = process.env.MESSAGE || 'Hello, World!';

// Connect to the RabbitMQ server
amqp.connect(queueAddress, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    const queue = 'output';

    channel.assertQueue(queue, {
      durable: true
    });

    // Function to send a message
    const sendMessage = () => {
      channel.sendToQueue(queue, Buffer.from(message));
      console.log("Sent: %s", message);
    };

    // Set interval for sending messages
    setInterval(sendMessage, interval);
  });
});
