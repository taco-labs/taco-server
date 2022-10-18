table "user_session" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "user_id" {
    type = uuid
    null = false
  }

  column "expire_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.id,
    ]
  }

  index "user_id_uidx" {
    unique = true
    columns = [
      column.user_id,
    ]
  }

  foreign_key "user_session_fk" {
    columns = [
      column.user_id,
    ]

    ref_columns = [
      table.user.column.id,
    ]

    on_delete = CASCADE

    on_update = NO_ACTION
  }
}

table "driver_session" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "driver_id" {
    type = uuid
    null = false
  }

  column "activated" {
    type = bool
    null = false
  }

  column "expire_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.id,
    ]
  }

  index "driver_id_uidx" {
    unique = true
    columns = [
      column.driver_id,
    ]
  }

  foreign_key "driver_session_fk" {
    columns = [
      column.driver_id,
    ]

    ref_columns = [
      table.driver.column.id,
    ]

    on_delete = CASCADE

    on_update = NO_ACTION
  }
}
