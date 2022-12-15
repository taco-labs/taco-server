table "user_referral" {
  schema = schema.taco

  column "from_user_id" {
    type = uuid
    null = false
    comment = "피추천인"
  }

  column "to_user_id" {
    type = uuid
    null = false
    comment = "추천인"
  }

  column "reward_rate" {
    type = int
    null = false
  }

  column "current_reward" {
    type = int
    null = false
  }

  column "reward_limit" {
    type = int
    null = false
  }

  column "create_time" { 
    type = timestamp
    null = false
  }

  primary_key {
    columns = [
      column.from_user_id,
    ]
  }

  foreign_key "user_referral_from_user_id_fk" {
    columns = [
      column.from_user_id,
    ]
    ref_columns = [
      table.user.column.id,
    ]
    on_delete = CASCADE
    on_update = NO_ACTION
  }

  foreign_key "user_referral_to_user_id_fk" {
    columns = [
      column.to_user_id,
    ]
    ref_columns = [
      table.user.column.id,
    ]
    on_delete = CASCADE
    on_update = NO_ACTION
  }
}

table "driver_referral" {
  schema = schema.taco

  column "driver_id" {
    type = uuid
    null = false
  }

  column "reward_rate" {
    type = int
    null = false
  }

  column "current_reward" {
    type = int
    null = false
  }

  column "reward_limit" {
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
    ]
  }

  foreign_key "driver_referral_driver_id_fk" {
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
