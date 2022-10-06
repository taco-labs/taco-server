enum "taxi_call_state" {
  schema = schema.taco

  values = [
    "TAXI_CALL_REQUESTED",
    "DRIVER_TO_DEPARTURE",
    "DRIVER_TO_ARRIVAL",
    "TAXI_CALL_DONE",
    "USER_CANCELLED",
    "DRIVER_CANCELLED",
    "TAXI_CALL_FAILED",
    "DRIVER_SETTLEMENT_DONE",
  ]
}

table "taxi_call_request" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "user_id" {
    type = uuid
    null = false
  }

  column "driver_id" {
    type = uuid
    null = true
  }

  column "departure" {
    type = jsonb
    null = false
  }

  column "arrival" {
    type = jsonb
    null = false
  }

  column "taxi_call_state" {
    type = enum.taxi_call_state
    null = false
  }

  column "payment_summary" {
    type = jsonb
    null = false
    comment = "User payment id / company / redacted card number"
  }

  // price (requested)
  column "request_base_price" {
    type = int
    null = false
  }

  column "request_min_additional_price"  {
    type = int
    null = false
  }

  column "request_max_additional_price" {
    type = int
    null = false
  }

  // price (actual)
  column "base_price" {
    type = int
    null = true
  }

  column "additional_price" {
    type = int
    null = true
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
      column.id,
    ]
  }

  index "taxi_call_request_user_id_idx" {
    unique = false
    type = HASH
    columns = [
      column.user_id,
    ]
  }

  index "taxi_call_request_driver_id_idx" {
    unique = false
    type = HASH
    columns = [
      column.driver_id,
    ]
  }

  foreign_key "user_taxi_call_request_fk" {
    columns = [
      column.user_id,
    ]

    ref_columns = [
      table.user.column.id,
    ]

    on_delete = NO_ACTION

    on_update = NO_ACTION
  }

  foreign_key "driver_taxi_call_request_fk" {
    columns = [
      column.driver_id,
    ]

    ref_columns = [
      table.driver.column.id,
    ]

    on_delete = NO_ACTION

    on_update = NO_ACTION
  }
}

table "taxi_call_ticket" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "taxi_call_request_id" {
    type = uuid
    null = false
  }

  column "attempt" {
    type = int
    null = false
  }

  column "additional_price" {
    type = int
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
      column.id,
    ]
  }

  index "taxi_call_ticket_taxl_call_request_id_idx" {
    unique = false
    type = HASH
    columns = [
      column.taxi_call_request_id,
    ]
  }

  foreign_key "taxi_call_request_id_fk" {
    columns = [
      column.taxi_call_request_id,
    ]
    
    ref_columns = [
      table.taxi_call_request.column.id,
    ]

    on_delete = CASCADE

    on_update = NO_ACTION
  }
}

table "taxi_call_last_received_ticket" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "taxi_call_ticket_id" {
    type = uuid
    null = false
  }

  column "receive_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
    ]
  }

  index "receive_time_brin_idx" {
    unique = false
    type = BRIN
    columns = [
      column.receive_time,
    ]
  }

  foreign_key "taxi_call_ticket_id_fk" {
    columns = [
      column.taxi_call_ticket_id,
    ]

    ref_columns = [
      table.taxi_call_ticket.column.id,
    ]

    on_delete = NO_ACTION
    on_update = NO_ACTION
  }

  foreign_key "driver_id_fk" {
    columns = [
      column.driver_id,
    ]

    ref_columns = [
      table.driver.column.id,
    ]

    on_delete = NO_ACTION
    on_update = NO_ACTION
  }
}

table "driver_location" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "location" {
    null = true
    type = sql("geometry(point,4326)")
  }

  column "on_duty" {
    type = boolean
    null = false
    comment = "Is taxi driver is activated (가입 승인을 받았는지 여부)"
  }

  primary_key {
    columns = [
      column.driver_id,
    ]
  }

  index "driver_location_idx" {
    unique = false
    type = GIST
    columns = [
      column.location,
    ]
  }

  foreign_key "driver_location_fk" {
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
