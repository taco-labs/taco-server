table "sms_verification" {
  schema = schema.taco

  column "id" {
    type = text
    null = false
  }

  column "verification_code" {
    type = text
    null = false
  }

  column "verified" {
    type = bool
    null = false
  }
  
  column "phone" {
    type = text
    null = false
  }

  column "expire_time" {
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.id,
    ]
  }
}
