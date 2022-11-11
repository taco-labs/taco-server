table "user_payment" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "user_id" {
    type = uuid
    null = false
  }

  column "name" {
    type = text
    null = false
  }

  column "card_company" {
    type = text
    null = false
  }

  column "redacted_card_number" {
    type = text
    null = false
  }

  column "card_expiration_year" {
    type = char(2)
    null =false
  }

  column "card_expiration_month" {
    type = char(2)
    null = false
  }

  column "billing_key" {
    type = text
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

  index "user_id_idx" {
    unique = false
    type = HASH
    columns = [
      column.user_id,
    ]
  }

  foreign_key "user_payment_fk" {
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

table "user_default_payment" {
  schema = schema.taco

  column "user_id" {
    type = uuid
    null = false
  }

  column "payment_id" {
    type = uuid
    null = false
  }

  primary_key {
    columns = [
      column.user_id,
    ]
  }

  index "user_default_payment_id_uidx" {
    unique = true
    columns = [
      column.payment_id,
    ]
  }

  foreign_key "user_default_payment_user_id_fk" {
    columns = [
      column.user_id,
    ]
    ref_columns = [
      table.user.column.id,
    ]
    on_delete = CASCADE
    on_update = NO_ACTION
  }

  foreign_key "user_default_payment_user_payment_fk" {
    columns = [
      column.payment_id,
    ]
    ref_columns = [
      table.user_payment.column.id,
    ]
    on_delete = CASCADE
    on_update = NO_ACTION
  }
}

table "user_payment_invalid_state" {
  schema = schema.taco

  column "payment_id" {
    type = uuid
    null = false
  }

  column "invalid_reason" {
    type = text
    null = false
  }

  column "invalid_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.payment_id,
    ]
  }

  foreign_key "user_payment_invalid_state_payment_fk" {
    columns = [
      column.payment_id,
    ]
    ref_columns = [
      table.user_payment.column.id,
    ]
    on_delete = CASCADE
    on_update = NO_ACTION
  }
}

table "user_payment_order" {
  schema = schema.taco

  column "user_id" {
    type = uuid
    null = false
  }

  column "order_id" {
    type = text
    null = false
  }

  column "payment_summary" {
    type = jsonb
    null = false
    comment = "User payment id / company / redacted card number"
  }

  column "order_name" {
    type = text
    null = false
  }

  column "price" {
    type = int
    null = false
  }

  column "payment_result" {
    type = jsonb
    null = true
    comment = "Payment result json from external payment service's response"
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
      column.order_id,
    ]
  }

  foreign_key "user_payment_order_user_id_fk" {
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

table "user_payment_failed_order" {
  schema = schema.taco

  column "user_id" {
    type = uuid
    null = false
  }

  column "order_id" {
    type = text
    null = false
  }

  column "order_name" {
    type = text
    null = false
  }

  column "price" {
    type = int
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.order_id,
    ]
  }

  foreign_key "user_payment_failed_order_user_id_fk" {
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
