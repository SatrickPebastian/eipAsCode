const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables
const [pipeAddressIn, pipeIn, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut, pipeTypeOut] = process.env.out.split(',');
const inRoutingKey = process.env.inRoutingKey || '#';
const outRoutingKey = process.env.outRoutingKey || '';

amqp.connect(pipeAddressIn, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    if (pipeTypeIn === 'queue') {
      channel.assertQueue(pipeIn);
      console.log("Waiting for messages in queue %s.", pipeIn);

      channel.consume(pipeIn, function(msg) {
        if (msg !== null) {
          const message = JSON.parse(msg.content.toString());
          console.log("Received: %s", JSON.stringify(message));

          forwardMessage(channel, pipeOut, pipeTypeOut, outRoutingKey, message);
          channel.ack(msg);
        }
      }, {
        noAck: false
      });

    } else if (pipeTypeIn === 'topic') {
      channel.assertExchange(pipeIn, 'topic');
      console.log("Waiting for messages on topic exchange %s with routing key %s.", pipeIn, inRoutingKey);

      //assert temporary queue
      channel.assertQueue('', { exclusive: true }, function(error2, q) {
        if (error2) {
          throw error2;
        }

        channel.bindQueue(q.queue, pipeIn, inRoutingKey);
        channel.consume(q.queue, function(msg) {
          if (msg !== null) {
            const message = JSON.parse(msg.content.toString());
            console.log("Received: %s", JSON.stringify(message));

            forwardMessage(channel, pipeOut, pipeTypeOut, outRoutingKey, message);
            channel.ack(msg);
          }
        }, {
          noAck: false
        });
      });

    } else {
      console.error(`Unknown pipe type for input: ${pipeTypeIn}`);
    }
  });
});

function forwardMessage(channel, pipeOut, pipeTypeOut, routingKey, message) {
  if (pipeTypeOut === 'queue') {
    channel.assertQueue(pipeOut);
    channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
    console.log("Forwarded message to queue %s", pipeOut);

  } else if (pipeTypeOut === 'topic') {
    channel.assertExchange(pipeOut, 'topic');
    channel.publish(pipeOut, routingKey, Buffer.from(JSON.stringify(message)));
    console.log("Forwarded message to topic exchange %s with routing key %s", pipeOut, routingKey);

  } else {
    console.error(`Unknown pipe type for output: ${pipeTypeOut}`);
  }
}
