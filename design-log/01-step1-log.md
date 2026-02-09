# Design Log: Step 1 - High-Level Architecture

## Decisions Made

### 1. The Core Event Backbone: Kafka vs RabbitMQ vs Pub/Sub
**Decision:** Use **Kafka (Confluent Cloud / MSK)** for the core transaction path.
**Reasoning:**
*   **Ordering:** We need strict ordering for order lifecycle (Created -> Validated -> Paid -> Kitchen -> Delivered). Kafka's partition key strategy is superior here.
*   **Durability:** Orders are financial transactions. We need a durable log that can be replayed if downstream consumers (like the Kitchen System) crash or need to re-process data. RabbitMQ's default ephemeral nature is risky here.
*   **Scale:** McDonald's scale implies high throughput burst traffic. Kafka handles backpressure natively by storing messages on disk; RabbitMQ can choke if consumers are slow and queues fill up RAM.

### 2. Extension Layer: GCP Pub/Sub
**Decision:** Bridge to **GCP Pub/Sub** for analytics and auxiliary services.
**Reasoning:**
*   **Serverless Fit:** Cloud Functions and Cloud Run work natively with Pub/Sub push subscriptions. This is much easier for the Data Science team than managing Kafka consumers.
*   **Isolation:** By using Kafka Connect to replicate to Pub/Sub, we isolate the "messy" read patterns of analytics from the mission-critical "write" path of order fulfillment.

### 3. Async Boundary at Ingestion
**Decision:** Immediate HTTP 202 Ack after producing to Kafka.
**Reasoning:**
*   **Latency:** The mobile app shouldn't wait for the inventory check. It just needs "Order Received".
*   **Resilience:** If the Inventory Service is down, we still accept the order. We can process it later when the service recovers.

## Pending Questions / Next Steps
*   **Schema Registry:** We haven't defined the payload structure yet. This is Step 2.
*   **Partition Strategy:** We need to ensure we don't have "hot partitions" if we key by StoreID (some stores are busier). Maybe key by OrderID?
