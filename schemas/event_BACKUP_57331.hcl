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

  index "event_create_time_event_uri_idx" {
    unique = false
    type = BRIN
    columns = [
      column.create_time,
    ]
  }
}
