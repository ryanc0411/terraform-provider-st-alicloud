terraform {
  required_providers {
    st-alicloud = {
      source = "example.local/myklst/st-alicloud"
    }
  }
}

provider "st-alicloud" {
  region = "ap-southeast-1"
}

resource "st-alicloud_ddoscoo_web_ai_protect_config" "test" {
  domain  = "api.1aa9t.com"
  aimode = "watch"
  aitemplate = "level60"
}
