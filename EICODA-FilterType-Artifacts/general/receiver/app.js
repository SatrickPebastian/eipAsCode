const amqp = require('amqplib/callback_api');
const process = require('process');

const [pipeAddress, pipe, pipeType] = process.env.in.split(',');
const routingKey = process.env.topicKey;

amqp.connect(pipeAddress, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    if (pipeType === 'queue') {
      channel.assertQueue(pipe, { durable: true });
      console.log("Waiting for messages in queue %s.", pipe);

      channel.consume(pipe, function(msg) {
        if (msg !== null) {
          console.log("Received: %s", msg.content.toString());
          channel.ack(msg);
        }
      });
    } else if (pipeType === 'topic') {
      channel.assertExchange(pipe, 'topic', { durable: true });
      console.log("Waiting for messages on topic %s with routing key %s.", pipe, routingKey);

      //temp queue bound to this exchange
      channel.assertQueue('', { exclusive: true }, function(error2, q) {
        if (error2) {
          throw error2;
        }

        channel.bindQueue(q.queue, pipe, routingKey);
        channel.consume(q.queue, function(msg) {
          if (msg !== null) {
            console.log("Received: %s", msg.content.toString());
            channel.ack(msg);
          }
        });
      });
    } else {
      console.error(`Unknown pipe type: ${pipeType}`);
    }
  });
});
