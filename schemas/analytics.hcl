// TODO (taekyeom) partitioning?
table "analytics" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "event_type" {
    type = text
    null = false
  }

  column "payload" {
    type = jsonb
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.id,
    ]
  }

  index "analytics_create_time_brin_idx" {
    unique = false
    type = BRIN
    columns = [
      column.create_time,
    ]
  }
}
