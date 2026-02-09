# Design Log: Step 3 - Streaming & Messaging Layer

## Decisions Made

### 1. Partition Key: `order_id` vs `store_id`
**Decision:** `order_id` (UUID).
**Reasoning:**
*   **Uniformity:** `store_id` causes hot spots (Times Square vs Rural). `order_id` ensures perfect distribution.
*   **Trade-off:** We lose ordering *per store* in the raw topic. However, downstream consumers (Analytics) can re-aggregate by `store_id` if needed. For the transactional path (Kitchen), `order_id` ordering is sufficient (we just need the sequence of events for *one* order).

### 2. `min.insync.replicas` = 2
**Decision:** Strict durability.
**Reasoning:**
*   With `rf=3`, `min.insync=1` (default) allows acking after just the leader writes. If the leader crashes before replicating, data is lost.
*   `min.insync=2` forces replication to a follower before ack. This adds ~2-5ms latency (cross-AZ network hop) but guarantees durability. For financial transactions (Orders), durability > latency.

### 3. Topic Retention: 7 Days
**Decision:** Long retention.
**Reasoning:**
*   Storage is cheap (S3 tiered storage in Confluent Cloud / MSK).
*   Operational recovery is expensive. Being able to replay a week of data saves engineering teams from manual database surgery during incidents.

## Pending Questions / Next Steps
*   **Consumer Group Sizing:** We have 30 partitions. Do we run 30 pods immediately?
    *   *Plan:* Run 3 pods (handling 10 partitions each) initially. Scale up based on `consumer_lag` metric.

## Next Step: Producers & Consumers
Building the actual Java/Go application logic that writes to and reads from these topics.
