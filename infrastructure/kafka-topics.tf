# Kafka Topic Configuration
# Provider: hashicorp/kafka or similar

resource "kafka_topic" "global_orders" {
  name               = "global.orders.v1"
  replication_factor = 3
  partitions         = 30

  config = {
    "cleanup.policy"        = "delete"
    "retention.ms"          = "604800000" # 7 days
    "segment.ms"            = "604800000"
    "min.insync.replicas"   = "2"         # Durability guarantee: need 2/3 brokers to ack
    "compression.type"      = "snappy"    # Balanced CPU/Bandwidth
    "message.timestamp.type"= "CreateTime"
  }
}

# Dead Letter Topic for poison pills
resource "kafka_topic" "global_orders_dlt" {
  name               = "global.orders.v1.dlt"
  replication_factor = 3
  partitions         = 5 # Lower volume expected

  config = {
    "cleanup.policy"      = "delete"
    "retention.ms"        = "259200000" # 3 days (shorter retention for errors)
    "min.insync.replicas" = "2"
  }
}
