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

  index "active_taxi_call_idx" {
    type = HASH
    columns = [
      column.id
    ]
    where = "taxi_call_state IN ('TAXI_CALL_REQUESTED', 'DRIVER_TO_DEPARTURE', 'DRIVER_TO_ARRIVAL')"
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

  column "rejected" {
    type = boolean
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

table "driver_taxi_call_context" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "on_duty" {
    type = boolean
    null = false
    comment = "Is taxi driver is activated (가입 승인을 받았는지 여부)"
  }

  column "can_recieve" {
    type = boolean
    null = false
    comment = "현재 ticket을 수신 가능한 상태인지 (eg. 주행 중일 때 false)"
  }

  column "last_received_request_ticket" {
    type = uuid
    null = false
  }

  column "receive_time" {
    type = timestamp
    null = false
  }

  index "driver_context_receive_time_brin_idx" {
    unique = false
    type = BRIN
    columns = [
      column.receive_time,
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

table "driver_taxi_call_settlement" {
  schema = schema.taco

  column "taxi_call_request_id" {
    type = uuid
    null = false
  }

  column "settlement_done" {
    type = boolean
    null = false
  }

  column "settlement_done_time" {
    type = timestamp
    null = false
  }

  foreign_key "settlement_taxi_call_request_id_fk" {
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
