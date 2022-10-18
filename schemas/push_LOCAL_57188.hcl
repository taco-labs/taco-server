table "push_token" {
  schema = schema.taco

  column "principal_id" {
    type = uuid
    null = false
  }

  column "fcm_token" {
    type = text
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  column "update_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.principal_id,
    ]
  }
}
