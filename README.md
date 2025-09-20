# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.

## Running the code

- Use `./rabbit.sh start/stop/logs` to manage the RabbitMQ container.
    - After running `./rabbit.sh start`, the RabbitMQ UI will be available at [http://localhost:15672](http://localhost:15672).
    - The default username and password are `guest` and `guest`.
- Run `go run cmd/client/main.go` to start the client.
- Run `go run cmd/server/main.go` to start the server.

## Notes about RabbitMQ

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

**Subscribers & Routing**

**Consumers** are programs that connect to queues and pull the messages out of them.

A queue can have 0, 1, or many consumers.

- If a queue has no consumers, messages will accumulate in the queue and never be processed.
- If a queue has one consumer, that consumer will process all messages in the queue (assuming it can keep up).
- If a queue has many consumers, messages will be distributed between them in a round-robin fashion (unless you set a priority).

The exclusive flag can be used to tell the RabbitMQ server to only allow one consumer to connect to the queue at a time. Often pub/sub needs fall into one of two categories:

- Process an event once-per-server-instance (good for ephemeral, exclusive queues with one consumer)
- Process an event once, period (good for durable, non-exclusive queues with many consumers)

Routing keys in RabbitMQ are made up of words separated by dots. For example, the routing key `user.created` is made up of two words: user and created.

RabbitMQ supports two types of **wildcards** in routing keys:

- `*` (star) substitutes for exactly one word
- `#` (hash) substitutes for zero or more words

Technically, you can name your exchanges, queues, and routing keys whatever you want, but it's critically important to choose good names. Not only will it make your system easier to understand, but it can also make it more flexible and powerful.

**Exchange Naming**: It's common for one "system" to all use the same exchange. Similar to how you might have a single "database" within a Postgres instance, you might just have a single "exchange" within a RabbitMQ instance.

**Queue Naming**: When I'm working with a direct key -> queue relationship, I'll often name the queue the same as the key, but add a word to describe the intended consumer. For example, if I have a routing key user.created, I might create a queue for my "email notifier" service called user.created.email_notifier.

If I have a queue that consumes all user events, I might name it user.all.billing_service.

If I have temporary queues, I might append a UUID to the queue name to ensure uniqueness. For example, maybe I have web servers that scale up and down based on traffic, and each server needs a copy of "comment created" events. I might name each server's queue one of:

comment.created.bb7a488b-b4e9-4b16-a697-51c20a09b87b
comment.created.8d0a9d3e-5244-460b-bacc-80ae2b802677
comment.created.6814c13f-c33b-4ff7-a4f8-98c718fea980
...

I'll often use auto-generated queue names like this with transient, auto-delete, and exclusive properties so they can be created and destroyed as the system restarts and scales.

**Routing Key Naming**: This is the one that you really want to get right. Not only do you want the routing key names to be descriptive, but you also want them to be flexible for potential wildcard matching. I've found that often a noun.verb pattern works well. For example:

user.created
user.updated
comment.created
comment.deleted
etc.

This allows you to easily bind queues to all events of a certain type, or all events that affect a certain entity.