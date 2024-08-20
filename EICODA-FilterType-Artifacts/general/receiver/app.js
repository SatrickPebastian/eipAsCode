const amqp = require('amqplib/callback_api');
const process = require('process');

const [queueAddress, queue] = process.env.in.split(',');

amqp.connect(queueAddress, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    channel.assertQueue(queue);

    console.log("Waiting for messages in %s.", queue);

    channel.consume(queue, function(msg) {
      if (msg !== null) {
        console.log("Received: %s", msg.content.toString());
        channel.ack(msg);
      }
    });
  });
});
