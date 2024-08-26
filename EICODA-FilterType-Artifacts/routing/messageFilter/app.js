const amqp = require('amqplib/callback_api');
const fs = require('fs');
const path = require('path');

const [pipeAddressIn, pipeIn, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut, pipeTypeOut] = process.env.out.split(',');
const inRoutingKey = process.env.inRoutingKey || '#';
const outRoutingKey = process.env.outRoutingKey || '';

const criteriaPath = '/etc/config/criteria';
const filterLogic = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

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
        handleIncomingMessage(channel, msg);
      }, {
        noAck: false
      });

    } else if (pipeTypeIn === 'topic') {
      channel.assertExchange(pipeIn, 'topic');
      console.log("Waiting for messages on topic exchange %s with topic key %s.", pipeIn, inRoutingKey);

      // asserts temporary queue for binding to exchange
      channel.assertQueue('', { exclusive: true }, function(error2, q) {
        if (error2) {
          throw error2;
        }

        channel.bindQueue(q.queue, pipeIn, inRoutingKey);
        channel.consume(q.queue, function(msg) {
          handleIncomingMessage(channel, msg);
        }, {
          noAck: false
        });
      });

    } else {
      console.error(`Unknown pipe type for input: ${pipeTypeIn}`);
    }
  });
});

function handleIncomingMessage(channel, msg) {
  const message = JSON.parse(msg.content.toString());

  if (filterMessage(message)) {
    if (pipeTypeOut === 'queue') {
      channel.assertQueue(pipeOut);
      channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
      console.log("Sent filtered message to queue %s: %s", pipeOut, JSON.stringify(message));

    } else if (pipeTypeOut === 'topic') {
      channel.assertExchange(pipeOut, 'topic');
      channel.publish(pipeOut, outRoutingKey, Buffer.from(JSON.stringify(message)));
      console.log("Sent filtered message to topic exchange %s with topic key %s: %s", pipeOut, outRoutingKey, JSON.stringify(message));

    } else {
      console.error(`Unknown pipe type for output: ${pipeTypeOut}`);
    }
  } else {
    console.log("Message filtered out: %s", JSON.stringify(message));
  }

  channel.ack(msg);
}

function filterMessage(message) {
  return filterLogic.criterias.every(rule => eval(rule.condition));
}
