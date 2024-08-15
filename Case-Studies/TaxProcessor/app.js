const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables

const [queueAddress, queue] = process.env.in.split(',');
console.log("queueAddress:"+queueAddress)
console.log("queueName:"+queue)

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

    console.log("Waiting for messages in %s. To exit press CTRL+C", queue);

    channel.consume(queue, function(msg) {
      if (msg !== null) {
        console.log("Received: %s", msg.content.toString());
        channel.ack(msg);
      }
    });
  });
});
