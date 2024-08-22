const amqp = require('amqplib/callback_api');

const [pipeAddressIn, pipeIn, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut, pipeTypeOut] = process.env.out.split(',');
const inRoutingKey = process.env.inRoutingKey || '#';
const outRoutingKey = process.env.outRoutingKey || '';
const dataToFilter = process.env.data.split(',');

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

        dataToFilter.forEach(field => {
          deleteNestedField(message.data, field.split('.'));
        });

        sendToOutputPipe(channel, pipeOut, pipeTypeOut, message);
        console.log("Sent filtered message to %s: %s", pipeOut, JSON.stringify(message));

        channel.ack(msg);
      }, {
        noAck: false
      });
    });
  });
});

//sets up input pipe (queue or topic)
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

//sets up output pipe (queue or topic)
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

//nested field deletion from object
function deleteNestedField(obj, fieldPath) {
  if (!obj) return;

  if (fieldPath.length === 1) {
    delete obj[fieldPath[0]];
  } else {
    const field = fieldPath.shift();
    if (obj[field]) {
      deleteNestedField(obj[field], fieldPath);
    }
  }
}
