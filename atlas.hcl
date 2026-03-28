variable "database_url" {
  type    = string
  default = getenv("DATABASE_URL")
}

env "dev" {
  src = "file://atlas/schema.hcl"
  
  # Development database URL from environment
  url = var.database_url
  
  # Migration directory
  migration {
    dir = "file://atlas/migrations"
  }
  
  # Diff policy
  diff {
    skip {
      drop_schema = true
      drop_table  = false
    }
  }
}

env "local" {
  src = "file://atlas/schema.hcl"
  
  # Local development database
  url = "postgres://cs_pg_user:cs_pg_password@localhost:5680/main-server?sslmode=disable"
  
  # Migration directory
  migration {
    dir = "file://atlas/migrations"
  }
  
  # Diff policy
  diff {
    skip {
      drop_schema = true
      drop_table  = false
    }
  }
}
