# Step 2: Event & Schema Design

## 2.1 Objective
Define the "Language of the Mesh". We must design a contract that allows producers and consumers to evolve independently. We chose **Apache Avro** with a strict **Forward Compatibility** policy.

## 2.2 The Schema: `OrderCreated.avsc`

The core event represents an order being placed. Note the use of **strict types** and **documentation**.

### Key Design Choices:
*   **Namespace `com.mcdonalds.orders`**: Prevents name collisions in the registry.
*   **`price_cents` (int)**: Avoids floating point errors (IEEE 754) common in financial systems. A price of `$4.99` is stored as `499`.
*   **`logicalType: timestamp-millis`**: Ensures all consumers interpret time consistently as epoch millis, not strings or custom formats.

```json
{
  "namespace": "com.mcdonalds.orders",
  "type": "record",
  "name": "OrderCreated",
  "fields": [
    {"name": "order_id", "type": "string"},
    {"name": "store_id", "type": "string"},
    {"name": "total_amount_cents", "type": "int"},
    {"name": "status", "type": {"type": "enum", "name": "OrderStatus", "symbols": ["NEW", "PAID", "KITCHEN", "COMPLETED", "CANCELLED"]}},
    {
       "name": "items",
       "type": {"type": "array", "items": "OrderItem"} 
    }
  ]
}
```

## 2.3 Schema Governance & Evolution

We use a **Schema Registry** (e.g., Confluent Schema Registry or AWS Glue) to enforce rules.

### The Rule: Forward Compatibility
*   **Definition:** New schemas can be read by old consumers.
*   **Why:** We can upgrade the **Order Service** (Producer) to include a new field like `"curbside_pickup_spot": "A1"`.
*   **Impact:** The old **Kitchen Display System** (Consumer) doesn't know about `curbside_pickup_spot`. With Forward Compatibility, it simply **ignores** the new field and processes the rest of the order. The system doesn't break.
*   **Prohibited Changes:**
    *   Deleting a required field.
    *   Renaming a field.
    *   Changing a field's type (e.g., `int` to `string`).

## 2.4 Interview Talking Points

### The 30-Second Pitch
"I enforced strict data governance using **Avro and a Schema Registry**. By defining `OrderCreated` as a typed contract and enforcing **Forward Compatibility**, I ensured that upstream changes (like adding a new feature) never break downstream consumers—preventing the most common cause of outages in distributed systems."

### The 2-Minute Deep Dive
"In a text-based system (JSON), data quality issues are discovered at *runtime*—usually when a consumer crashes at 3 AM because a field name changed. I shifted this discovery to *build time*.

I chose **Avro** over Protobuf or JSON because of its compact binary format and first-class support in the Kafka ecosystem. I designed the schema with 'money' fields as integers (`price_cents`) to avoid floating-point drift, which is critical for financial reconciliation.

The most important architectural decision here was the **Compatibility Policy**. I enforced **Forward Compatibility**. This allows us to decouple the release cycles of our microservices. The Order Team can ship features on Tuesday, and the Reporting Team can update their consumers on Friday, without coordination meetings."
