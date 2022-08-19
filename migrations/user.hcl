table "user" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "first_name" {
    type = text
    null = false
  }

  column "last_name" {
    type = text
    null = false
  }

  column "email" {
    type = text
    null = false
  }

  column "birthday" {
    type = text
    null = false
  }

  column "phone" {
    type = text
    null = false
  }

  column "gender" {
    type = text
    null = false
  }

  column "app_os" {
    type = enum.app_os
    null = false
  }

  column "app_version" {
    type = text
    null = false
  }

  column "app_fcm_token" {
    type = text
    null = false
  }

  column "user_unique_key" {
    type = text
    comment = "User unique key from authentication service (eg. IamPort)"
    null = false
  }

  column "default_payment_id" {
    type = uuid 
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

  column "delete_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.id,
    ]
  }

  index "user_unique_key_uidx" {
    unique = true
    columns = [
      column.user_unique_key,
    ]
  }

  foreign_key "user_default_payment_fk" {
    columns = [
      column.default_payment_id,
    ]

    ref_columns = [
      table.user_payment.column.id,
    ]

    on_delete = NO_ACTION 

    on_update = NO_ACTION
  }
}

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

  column "card_number" {
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

  column "cardPassword" {
    type = char(2)
    null = false
  }

  column "customer_identity_number" {
    type = text
    null = false
  }

  column "create_time" {
    type = timestamp
    null = false
  }

  column "delete_time" {
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
