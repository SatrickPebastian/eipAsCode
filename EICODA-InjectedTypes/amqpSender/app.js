const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables
const interval = parseInt(process.env.interval, 10) || 5000;
const [queueAddress, queue] = process.env.outputPipe.split(',');
const message = process.env.message || 'Hello, World!';

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
      channel.sendToQueue(queue, Buffer.from(message));
      console.log("Sent: %s", message);
    };

    // Set interval for sending messages
    setInterval(sendMessage, interval);
  });
});
