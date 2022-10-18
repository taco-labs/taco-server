table "spatial_ref_sys" {
  schema = schema.public
  column "srid" {
    null = false
    type = integer
  }
  column "auth_name" {
    null = true
    type = character_varying(256)
  }
  column "auth_srid" {
    null = true
    type = integer
  }
  column "srtext" {
    null = true
    type = character_varying(2048)
  }
  column "proj4text" {
    null = true
    type = character_varying(2048)
  }
  primary_key {
    columns = [column.srid]
  }
  check "spatial_ref_sys_srid_check" {
    expr = "((srid > 0) AND (srid <= 998999))"
  }
}
schema "public" {
}
