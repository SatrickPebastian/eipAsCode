const amqp = require('amqplib/callback_api');

const [pipeAddressIn, pipeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut] = process.env.out.split(',');
const dataToFilter = process.env.data.split(',');

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

      //apply filter
      dataToFilter.forEach(field => {
        deleteNestedField(message.data, field.split('.'));
      });

      channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
      console.log("Sent filtered message to %s: %s", pipeOut, JSON.stringify(message));

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});

//deletes a nested field
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
