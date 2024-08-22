const amqp = require('amqplib/callback_api');
const fs = require('fs');
const jsonata = require('jsonata');

const [pipeAddressIn, pipeIn, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut, pipeTypeOut] = process.env.out.split(',');
const inRoutingKey = process.env.inRoutingKey || '#';
const outRoutingKey = process.env.outRoutingKey || '';

const criteriaPath = '/etc/config/criteria';
const transformationLogic = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

amqp.connect(pipeAddressIn, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    setupInputPipe(channel, pipeIn, pipeTypeIn, function(inputQueue) {
      setupOutputPipe(channel, pipeOut, pipeTypeOut);

      console.log("Waiting for messages in %s.", inputQueue);

      channel.consume(inputQueue, function(msg) {
        const message = JSON.parse(msg.content.toString());

        // apply jsonata translation to the entire message object
        const transformedData = {};
        for (let [key, expr] of Object.entries(transformationLogic)) {
          const expression = jsonata(expr);
          transformedData[key] = expression.evaluate(message); 
        }

        // merge translated data back into consumed message
        const translatedMessage = {
          ...message, 
          data: transformedData
        };

        sendToOutputPipe(channel, pipeOut, pipeTypeOut, translatedMessage);
        console.log("Sent translated message to %s: %s", pipeOut, JSON.stringify(translatedMessage));

        channel.ack(msg);
      }, {
        noAck: false
      });
    });
  });
});

function setupInputPipe(channel, pipeIn, pipeTypeIn, callback) {
  if (pipeTypeIn === 'queue') {
    channel.assertQueue(pipeIn);
    callback(pipeIn);
  } else if (pipeTypeIn === 'topic') {
    channel.assertExchange(pipeIn, 'topic');
    channel.assertQueue('', { exclusive: true }, function(error2, q) {
      if (error2) {
        throw error2;
      }
      channel.bindQueue(q.queue, pipeIn, inRoutingKey);
      callback(q.queue);
    });
  } else {
    console.error(`Unknown input pipe type: ${pipeTypeIn}`);
  }
}

function setupOutputPipe(channel, pipeOut, pipeTypeOut) {
  if (pipeTypeOut === 'queue') {
    channel.assertQueue(pipeOut);
  } else if (pipeTypeOut === 'topic') {
    channel.assertExchange(pipeOut, 'topic');
  } else {
    console.error(`Unknown output pipe type: ${pipeTypeOut}`);
  }
}

//sends message to the correct output pipe
function sendToOutputPipe(channel, pipeOut, pipeTypeOut, message) {
  if (pipeTypeOut === 'queue') {
    channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
    console.log("Sent message to queue %s: %s", pipeOut, JSON.stringify(message));
  } else if (pipeTypeOut === 'topic') {
    channel.publish(pipeOut, outRoutingKey, Buffer.from(JSON.stringify(message)));
    console.log("Sent message to topic exchange %s with routing key %s: %s", pipeOut, outRoutingKey, JSON.stringify(message));
  } else {
    console.error("Unknown pipe type for output.");
  }
}
