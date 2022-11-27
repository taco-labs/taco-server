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

  column "tag_ids" {
    type = sql("int[]")
    null = false
  }

  column "user_tag" {
    type = varchar(10)
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

  index "create_time_brin_idx" {
    unique = false
    type = BRIN
    columns = [
      column.create_time,
    ]
  }
}

table "taxi_call_ticket" {
  schema = schema.taco

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

  column "ticket_id" {
    type = uuid
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.taxi_call_request_id,
      column.attempt,
      column.additional_price,
    ]
  }

  index "taxi_call_ticket_id_uidx" {
    type = HASH
    columns = [
      column.ticket_id,
    ]
  }

  index "ticket_create_time_brin_idx" {
    unique = false
    type = BRIN
    columns = [
      column.create_time,
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

table "driver_taxi_call_context" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "can_receive" {
    type = boolean
    null = false
    comment = "현재 ticket을 수신 가능한 상태인지 (eg. 주행 중일 때 false)"
  }

  column "last_received_request_ticket" {
    type = uuid
    null = false
  }

  column "rejected_last_request_ticket" {
    type = boolean
    null = false
  }

  column "last_receive_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
    ]
  }

  index "driver_context_receive_time_brin_idx" {
    unique = false
    type = BRIN
    columns = [
      column.last_receive_time,
    ]
  }

  index "last_received_request_ticket_idx" {
    unique = false
    columns = [
      column.last_received_request_ticket,
    ]
  }

  foreign_key "driver_id_fk" {
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

  column "update_time" {
    type = timestamp
    null = false
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
