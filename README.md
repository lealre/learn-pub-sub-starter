# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.

## Notes

Pub/Sub (publish/subscribe) is a pattern software systems can use to communicate. Are particularly useful when many events are happening in real-time and disparate parts of the system need to react to those events.

Publishers publish, and subscribers subscribe. Alternative terms:

- Publisher = Producer = Sender
- Subscriber = Consumer = Receiver

Are often used to enable "event-driven design", or "event-driven architecture". An event-driven architecture uses events to trigger and communicate between decoupled systems.

**Message Broker**

A **message broker** is a middleman that allows different parts of the system to communicate without knowing about each other.

**RabbitMQ** is a popular open-source message broker that implements the **AMQP protocol**. It's open-source, flexible, powerful, and (reasonably) easy to use.

To run a RabbitMQ container, run the following command:

```bash
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
```

The **Advanced Message Queuing Protocol (AMQP)** is an open standard for passing business messages between applications or organizations.

While Kafka, SQS, Google Pub/Sub, and NATS use their own protocols (not AMQP), AMQP is still an open standard, and other message brokers like ActiveMQ, Qpid, Solace and Azure Service Bus also implement it.

Two other popular protocols are:

- MQTT is designed for small IoT devices and as such is optimized to be lightweight and energy efficient.
- STOMP is designed for web applications and is designed to be simple and easy to use.

**Publishers and Queues**

In RabbitMQ, an [**exchange**](https://www.rabbitmq.com/tutorials/amqp-concepts#exchanges) is where publishers send messages, typically with a routing key.

The exchange takes the message, uses the routing key as a filter, and sends the message to any queues that are listening for that routing key.

Publishers don't know about queues at all. They just send messages to exchanges, sometimes with a routing key.

Terminology:

- Exchange: A routing agent that sends messages to queues.
- Binding: A link between an exchange and a queue that uses a routing key to decide which messages go to the queue.
- Queue: A buffer in the RabbitMQ server that holds messages until they are consumed.
- Channel: A virtual connection inside a connection that allows you to create queues, exchanges, and publish messages.
- Connection: A TCP connection to the RabbitMQ server.

RabbitMQ Exchange Types:

- Direct: Messages are routed to the queues based on the message routing key exactly matching the binding key of the queue.
- Topic: Messages are routed to queues based on wildcard matches between the routing key and the routing pattern specified in the binding.
- Fanout: It routes messages to all of the queues bound to it, ignoring the routing key.
- Headers: Routes based on header values instead of the routing key. It's similar to topic but uses message header attributes for routing.

direct and topic are the most commonly useful in backend Pub/Sub architectures.

**Queues** are where the messages are stored after being routed through the exchange. Messages sit in a queue until they are consumed by a subscriber.

Queues can be "durable" or "transient". Durable queues survive a RabbitMQ server restart, while transient queues do not.

The metadata of a durable queue is stored on disk, while transient queues are only stored in memory.

We can also set the auto-delete and exclusive properties of our queues:

- Exclusive: The queue can only be used by the connection that created it.
- Auto-delete: The queue will be automatically deleted when its last connection is closed.
