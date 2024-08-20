const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables
const [pipeAddressIn, pipeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut] = process.env.out.split(',');

amqp.connect(pipeAddressIn, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    channel.assertQueue(pipeIn);
    channel.assertQueue(pipeOut);

    console.log("Waiting for messages in %s.", pipeIn);

    channel.consume(pipeIn, function(msg) {
      const message = JSON.parse(msg.content.toString());
      
      console.log("Received: %s", JSON.stringify(message));

      channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
      console.log("Forwarded message to %s", pipeOut);

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});
