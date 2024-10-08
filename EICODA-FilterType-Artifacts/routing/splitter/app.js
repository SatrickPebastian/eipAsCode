const amqp = require('amqplib/callback_api');
const process = require('process');

const [pipeAddressIn, pipeIn, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut, pipeTypeOut] = process.env.out.split(',');
const inRoutingKey = process.env.inRoutingKey || '#';
const outRoutingKey = process.env.outRoutingKey || '';
const dataToSplit = process.env.data.split(',');
const source = process.env.source;
const type = process.env.eventType;

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

        dataToSplit.forEach(fieldPath => {
          const fieldValue = getFieldValue(message, fieldPath);

          if (fieldValue !== undefined) {
            const fieldKey = fieldPath.split('.').pop();
            const cloudEventMessage = {
              specversion: '1.0',
              id: `id-${Math.random()}`,
              source: source,
              type: type,
              time: new Date().toISOString(),
              data: { [fieldKey]: fieldValue }
            };

            sendToOutputPipe(channel, pipeOut, pipeTypeOut, cloudEventMessage);
            console.log(`Sent: ${JSON.stringify(cloudEventMessage)} to ${pipeOut}`);
          } else {
            console.log(`Field "${fieldPath}" not found in the incoming message.`);
          }
        });

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

function getFieldValue(obj, fieldPath) {
  const adjustedPath = fieldPath.replace(/^message\./, '');
  return adjustedPath.split('.').reduce((o, key) => (o && o[key] !== undefined) ? o[key] : undefined, obj);
}
