# Step 3: Streaming & Messaging Layer - Configuration

## 3.1 Objective
Configure the "event highway" to support thousands of concurrent orders. The goal is linear scalability and regional fault tolerance.

## 3.2 Kafka Topic: `global.orders.v1`

### Configuration Details
*   **Partitions:** 30
    *   *Why?* Allows us to scale up to 30 concurrent consumers (e.g., 30 pods of the `Kitchen Service`).
*   **Replication Factor:** 3
    *   *Why?* Standard for high availability. 1 leader, 2 followers. Survives loss of 1 broker with zero impact; loss of 2 with potential unavailability but no data loss (if `min.insync` is met).
*   **Partitioner Strategy:** `DefaultPartitioner` (Murmur2 hash of key).
    *   *Key:* `order_id` (UUID).
    *   *Why?* Uniform distribution across all 30 partitions. Prevents "hot partitions" from busy stores.

### Reliability Settings
*   **`min.insync.replicas=2`**: The producer will only receive an ACK if the leader AND at least one follower have written the message to disk. This prevents "phantom writes" if a leader crashes immediately after acking.
*   **`retention.ms=7 days`**: Allows rewinding offsets to replay a full week of data in case of downstream logic bugs.

## 3.3 Dead Letter Topic (DLT)

*   **Topic:** `global.orders.v1.dlt`
*   **Purpose:** Events that fail schema validation or cause consumer crashes (poison pills) are moved here after 3 retries.
*   **SRE Workflow:** An alert fires on DLT ingress. On-call engineer inspects the payload, fixes the bug, and replays the message.

## 3.4 Interview Talking Points

### The 30-Second Pitch
"I designed the topic `global.orders.v1` with **30 partitions** backed by a **random UUID partition key** to eliminate hot-spotting. By configuring `min.insync.replicas=2`, I guaranteed that every order is written to at least two Availability Zones before we acknowledge the customer, trading a few milliseconds of latency for absolute data durability."

### The 2-Minute Deep Dive
"The biggest risk in partitioning is data skew. If I partitioned by `store_id`, a busy store during lunch rush would overwhelm a single partition and its consumer, creating a 'noisy neighbor' effect where latency spikes for everyone on that shard.

I chose to partition by `order_id` to achieve **uniform distribution**. This allows us to scale consumers linearly. If traffic doubles, we just double the consumer instances.

For retention, I set a **7-day policy**. This is a 'safety buffer'. If our Data Warehouse ETL job crashes on Friday night and nobody notices until Monday, we can simply rewind the offsets and replay the weekend's data without any loss. This turns a catastrophic data loss incident into a mundane operational task."
