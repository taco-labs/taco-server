table "event" {
  schema = schema.taco

  column "message_id" {
    type = uuid
    null = false
  }

  column "event_uri" {
    type = text
    null = false
  }

  column "delay_seconds" {
    type = bigint
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
      column.message_id,
    ]
  }

  index "event_uri_create_time_idx" {
    unique = false
    columns = [
      column.event_uri,
      column.create_time,
    ]
  }
}
