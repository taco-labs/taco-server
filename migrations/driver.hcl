enum "driver_type" {
  schema = schema.taco
  values = [
    "INDIVIDUAL",
    "COORPERATE",
  ]
}

table "driver" {
  schema = schema.taco

  column "id" {
    type = uuid
    null = false
  }

  column "driver_type" {
    type = enum.driver_type
    null = false
  }

  column "driver_on_duty" {
    type = boolean
    null = false
    comment = "택시 기사 출근 여부"
  }

  column "first_name" {
    type = text
    null = false
  }

  column "last_name" {
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

  column "driver_license_id" {
    type = text
    null = false
  }

  column "driver_license_image_url" {
    type = text
    null = false
    comment = "S3 URL for driver license card"
  }

  column "user_unique_key" {
    type = text
    comment = "User unique key from authentication service (eg. IamPort)"
    null = false
  }

  column "active" {
    type = boolean
    null = false
    comment = "Is taxi driver is activated (가입 승인을 받았는지 여부)"
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

  index "driver_user_unique_key_uidx" {
    unique = true
    columns = [
      column.user_unique_key,
    ]
  }
}

table "driver_settlement_account" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "name" {
    type = text
    null = false
  }

  column "bank" {
    type = text
    null = false
    comment = "Bank enum value"
  }

  column "accountNumber" {
    type = text
    null = false
    comment = "계좌번호"
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
      column.driver_id,
    ]
  }

  foreign_key "driver_settlement_account_fk" {
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
