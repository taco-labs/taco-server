enum "settlement_transfer_process_state"  {
  schema = schema.taco

  values = [
    "TRANSFER_REQUEST_RECEIVED",
    "TRANSFER_REQUESTED",
    "TRANSFER_EXECUTED",
  ]
}

table "driver_inflight_settlement_transfer" {
  schema = schema.taco

  column "transfer_id" {
    type = uuid
    null = false
  }

  column "driver_id" { 
    type = uuid
    null = false
  }

  column "execution_key" {
    type = text
    null = false
  }

  column "bank_transaction_id" {
    type = text
    null = false
  }

  column "amount" {
    type = int
    null = false
  }

  column "amount_without_tax" {
    type = int
    null = false
  }

  column "message" {
    type = varchar(10)
    null = false
  }

  column "state" {
    type = enum.settlement_transfer_process_state
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
      column.transfer_id,
    ]
  }

  index "inflight_settlement_transfer_driver_id_uidx" {
    unique = true
    columns = [
      column.driver_id,
    ]
  }
}

table "driver_failed_settlement_transfer" {
  schema = schema.taco

  column "transfer_id" {
    type = uuid
  }

  column "driver_id" { 
    type = uuid
    null = false
  }

  column "execution_key" {
    type = text
    null = false
  }

  column "bank_transaction_id" {
    type = text
    null = false
  }

  column "amount" {
    type = int
    null = false
  }

  column "amount_without_tax" {
    type = int
    null = false
  }

  column "message" {
    type = varchar(10)
    null = false
  }

  column "failure_message" {
    type = text
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.transfer_id,
    ]
  }
}

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
      column.driver_id,
      column.taxi_call_request_id,
      column.create_time,
    ]
  }

  /* foreign_key "settlement_taxi_call_request_id_fk" { */
  /*   columns = [ */
  /*     column.taxi_call_request_id, */
  /*   ] */
  /**/
  /*   ref_columns = [ */
  /*     table.taxi_call_request.column.id, */
  /*   ] */
  /**/
  /*   on_delete = CASCADE */
  /**/
  /*   on_update = NO_ACTION */
  /* } */
}

table "driver_total_settlement" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "total_amount" {
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

  column "amount" {
    type = int
    null = false
  }

  column "amount_without_tax" {
    type = int
    null = false
  }

  column "bank" {
    type = text
    null = false
    comment = "Bank enum value"
  }

  column "account_number" {
    type = text
    null = false
    comment = "계좌번호"
  }

  column "request_time" {
    type = timestamp
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
      column.create_time,
    ]
  }
}

table "driver_promotion_reward_limit" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "receive_count" {
    type = int
    null = false
  }

  column "receive_count_limit" {
    type = int
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
    ]
  }

  foreign_key "promotion_reward_limit_driver_id_fk" {
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

table "driver_promotion_reward_history" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "receive_date" {
    type = timestamp 
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
      column.receive_date,
    ]
  }

  foreign_key "promotion_reward_history_driver_id_fk" {
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

table "driver_promotion_settlement_reward" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "total_amount" {
    type = int
    null = false
  }

  primary_key {
    columns = [
      column.driver_id,
    ]
  }

  foreign_key "settlement_promotion_reward_driver_id_fk" {
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
