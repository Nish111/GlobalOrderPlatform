# Design Log: Step 2 - Event & Schema Design

## Decisions Made

### 1. Schema Format: Avro vs JSON vs Protobuf
**Decision:** **Apache Avro**.
**Reasoning:**
*   **Compactness:** Avro is binary and very small on the wire. Given our scale (thousands of orders/sec), saving 50% on bandwidth and storage costs millions. JSON is verbose.
*   **Ecosystem:** Kafka's Schema Registry has the best tooling for Avro. Protobuf is great for gRPC, but Avro is the standard for data-at-rest and event streaming.

### 2. Compatibility Mode: Forward vs Backward vs Full
**Decision:** **Forward Compatibility**.
**Reasoning:**
*   **Producer-First Workflow:** In our business, the "Order Taking" app evolves fastest (new promos, new delivery methods). We want to deploy the Producer *first*, and let Consumers catch up later.
*   **Safety:** Old consumers (creating vital daily reports) won't crash when receiving new data formats. They'll just miss the new fields until they upgrade.

### 3. Currency Handling
**Decision:** `int` for `price_cents`.
**Reasoning:**
*   Floating point math is notorious for errors like `0.1 + 0.2 = 0.30000000000000004`. In a financial system, these errors accumulate. Using integers for the smallest unit (cents) eliminates this class of bug entirely.

## Pending Questions / Next Steps
*   **Topic Naming:** Should the topic be `orders` or `orders-v1`?
    *   *Decision:* usage of `orders` is preferred, with schema versioning handling the evolution.
*   **Key Strategy:** What is the partition key?
    *   *Decision:* `order_id` (UUID). This ensures all events for a single order land on the same partition (guaranteed ordering).

## Next Step: Streaming & Messaging Layer
Now that we know *what* data looks like, we need to design *where* it goes. We need to configure the Kafka topics, partitions, and replication factors.
