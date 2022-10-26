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

  column "car_number" {
    type = text
    null = false
  }

  column "active" {
    type = boolean
    null = false
    comment = "Is taxi driver is activated (가입 승인을 받았는지 여부)"
  }

  column "on_duty" {
    type = boolean
    null = false
    comment = "Is taxi driver is on duty (출근 중인지 여부)"
  }

  column "driver_license_id" {
    type = text
    null = false
  }

  column "company_registration_number" {
    type = text
    null = false
  }

  column "driver_license_image_uploaded" {
    type = boolean
    null = false
    default = false
    comment = "Flag that driver license image uploaded"
  }

  column "driver_profile_image_uploaded" {
    type = boolean
    null = false
    default = false
    comment = "Flag that driver profile image uploaded"
  }

  column "user_unique_key" {
    type = text
    comment = "User unique key from authentication service (eg. IamPort)"
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

