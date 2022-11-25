table "driver_settlement_request" {
  schema = schema.taco

  column "taxi_call_request_id" {
    type = uuid
    null = false
  }

  column "driver_id" {
    type = uuid
    null = false
  }

  column "amount" {
    type = int
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.taxi_call_request_id,
    ]
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

table "driver_expected_settlement" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "expected_amount" {
    type = int
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
    ]
  }
}

table "driver_settlement_history" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "settlement_period_start" {
    type = timestamp
    null = false
  }

  column "settlement_period_end" {
    type = timestamp
    null = false
  }

  column "amount" {
    type = int
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
      column.settlement_period_start,
      column.settlement_period_end,
    ]
  }

  // TODO (taekyeom) Add brin index in create_time
}
