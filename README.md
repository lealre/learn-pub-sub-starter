# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.

## Notes

Pub/Sub (publish/subscribe) is a pattern software systems can use to communicate. Are particularly useful when many events are happening in real-time and disparate parts of the system need to react to those events.

Publishers publish, and subscribers subscribe. Alternative terms:

- Publisher = Producer = Sender
- Subscriber = Consumer = Receiver

Are often used to enable "event-driven design", or "event-driven architecture". An event-driven architecture uses events to trigger and communicate between decoupled systems.

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
