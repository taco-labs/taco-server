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

  column "birthday" {
    type = text
    null = false
  }

  column "gender" {
    type = text
    null = false
  }

  column "phone" {
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

  index "user_unique_key_uidx" {
    unique = true
    columns = [
      column.user_unique_key,
    ]
  }
}

